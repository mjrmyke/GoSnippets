//Myke Walker
//8-13-2017
//Proxy Program to act as an intermediary between a server and client
//Client sends an encrypted request to the TCP connection of the proxy, and the proxy
//decrypts it, and forwards it to its desired destination. The proxy also receives unencrypted data from the destination
//and encrypts it, then ships it back to the original client
//
// go run proxy.go -key=1234567890123456 -port=2222
//
// key flag is for encryption key (16,24,32 bytes)
// port flag is which port the proxy will listen on
// run program with -h for help with flags and to see defaults

package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

//Struct to hold simulated packet data
//all fields are []byte so they may be encrypted
type Packet struct {
	To      []byte `json: "To"`
	From    []byte `json: "From"`
	Message []byte `json: "Message"`
}

//channels to communicate to each thread
var packettoclient = make(chan Packet)
var packettodest = make(chan Packet)

//waitgroup to make sure main function ending does not end program
var mainwg sync.WaitGroup

//global secret key for encryption. Length of Key determines  16 = AES128, 24 = AES192, 32 = AES256
var AESkey = flag.String("key", "1234567890123456", "Encryption Key - 16, 24, or 32 byes")

func main() {

	//recv port for TCP server
	Rcvport := flag.String("port", "2222", "port to receive on")
	flag.Parse()

	//start the TCP server with the associated port
	go TCPProxyServer(*Rcvport)

	//have the waitgroup wait until notified that the server has closed
	mainwg.Add(1)
	mainwg.Wait()

}

func TCPProxyServer(port string) {
	defer mainwg.Done()

	//listen for TCP conns on the specified port
	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Awaiting TCP connection from Client")

	//sets the last the func will do before exiting, close the connections
	defer ln.Close()

	//for each connection, accept it, and then place it in its own thread
	for {
		connection, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection Error: ", err)
			break
		}
		fmt.Println("Accepted connection", connection)
		go HandleProxyConn(connection)

	}
	mainwg.Done()

}

func HandleProxyConn(conn net.Conn) {
	// add json decoder to connection
	d := json.NewDecoder(conn)

	var packet Packet

	fmt.Println("awaiting packet")

	///wait for a packet to be sent
	err := d.Decode(&packet)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("received packet")

	//decrypt packet header to determine where to pass the data along to
	decryptedmsg := decrypt(packet.Message)
	decrypteddest := decrypt(packet.To)
	// output message received
	fmt.Println("Packet Received:", string(packet.Message))
	fmt.Println("Decrypted Message: ", string(decryptedmsg))
	fmt.Println("Decrypted Destination: ", string(decrypteddest))
	fmt.Println("Dialing Destination")

	// connect to socket specified in packet
	destconn, err := net.Dial("tcp", "127.0.0.1:"+string(decrypteddest))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Destination Connected - Passing through to ProxyFunc")
	go HandleIncoming(conn)
	go HandleOutgoing(destconn, packet)
	//have the waitgroup wait until notified that the server has closed
	mainwg.Add(2)

}

//function to be ran in its own goroutine
//Sends and Receives information to Clients connection. Message is always encrypted
func HandleIncoming(conn net.Conn) {
	defer mainwg.Done()
	var err error
	var pckt Packet

	//if last iteration was successful, or if first run
	for err == nil {

		fmt.Println("waiting for packet")
		pckt = <-packettoclient
		fmt.Println("packet received!")

		//marshal the struct to json
		tmpdata, err := json.Marshal(pckt)
		if err != nil {
			log.Fatal(err)
		}

		// send to socket
		_, err = conn.Write(tmpdata)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("sent to Client")

		//attatch decoder to the connection and wait for the packet
		d := json.NewDecoder(conn)

		var packet Packet

		fmt.Println("awaiting packet")

		err = d.Decode(&packet)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("received")

		//once packet is received, pass it to the packettodest channel for shipping
		packettodest <- packet

	}
}

//function to be ran in its own goroutine
//Sends and Receives information to Server,
//information received is to be encrypted, packaged in a packet
//and shipped back to client via the packettoclient channel

func HandleOutgoing(conn net.Conn, packet Packet) {
	defer mainwg.Done()
	//handles base case
	firstrun := true

	//while loop
	for {

		//if not first run, wait for a packet to send
		if !firstrun {
			fmt.Println("Waiting for Packet from Client")
			packet = <-packettodest
		}

		//decrypt information
		tmpdecrypt := decrypt(packet.Message)
		fmt.Println("Encrypted Message Received: ", string(packet.Message))
		fmt.Println("Decrypted Message to be Sent: ", string(tmpdecrypt))

		//write the decrypted data to the connection of Server
		_, err := conn.Write(tmpdecrypt)
		if err != nil {
			log.Fatal(err)
		}

		// listen for reply from server
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		//encrypt, place in package, and send to channel
		EncryptShipPacket(message)

		//base case handled!
		firstrun = false
		fmt.Println("Sent to Server")
	}
}

//func that encrypts, places in package, and sends to channel
func EncryptShipPacket(message string) {
	msg := encrypt([]byte(message))
	pckt := Packet{Message: msg}
	packettoclient <- pckt
}

// func to create cipher using global AES key
func cipherCreation() cipher.Block {
	key := []byte(*AESkey)
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	return block
}

// func to encrypt input bytes
func encrypt(inputtext []byte) []byte {
	block := cipherCreation()

	//make initialization vector, and fill it randomly
	initvec := make([]byte, aes.BlockSize)
	_, err := io.ReadFull(rand.Reader, initvec)
	if err != nil {
		log.Fatal(err)
	}

	//stream the input into the cipher
	cipherinput := []byte(inputtext)
	stream := cipher.NewCTR(block, initvec)
	stream.XORKeyStream(cipherinput, inputtext)
	cipherinput = append(initvec, cipherinput...)

	return cipherinput
}

//func to decrypt input bytes
func decrypt(inputtext []byte) []byte {
	//create block
	block := cipherCreation()

	//size input text to work correctly
	initvec := inputtext[:aes.BlockSize]
	plaintext := make([]byte, len(inputtext)-aes.BlockSize)

	//decode
	stream := cipher.NewCTR(block, initvec)
	stream.XORKeyStream(plaintext, inputtext[aes.BlockSize:])

	return plaintext
}
