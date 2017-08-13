//Myke Walker
//8-13-2017
//TCP Server Program to receive a Client Connection
//Client sends a string, Server ToUppers the string, then returns
//
// go run server.go -key=1234567890123456 -port=3333
//
// key flag is for encryption key (16,24,32 bytes)
// port flag is which port the Server will listen on
// run program with -h for help with flags and to see defaults

package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

//Struct to hold simulated packet data
//all fields are []byte so they may be encrypted
type Packet struct {
	To      []byte `json: "To"`
	From    []byte `json: "From"`
	Message []byte `json: "Message"`
}

//waitgroup to make sure main function ending does not end program
var mainwg sync.WaitGroup

//global secret key for encryption. Length of Key determines  16 = AES128, 24 = AES192, 32 = AES256
var AESkey = flag.String("key", "1234567890123456", "Encryption Key - 16, 24, or 32 byes")

//entry point of application
func main() {

	//recv port for TCP server
	port := flag.String("port", "3333", "port to receive connections on")

	flag.Parse()

	fmt.Println("Server starting up")

	//start goroutine of the server
	go TCPServer(*port)

	// wait for the server to close
	mainwg.Add(1)
	mainwg.Wait()

}

//Function that starts a TCP server on the port specified by the argument when func is called
func TCPServer(port string) {
	defer mainwg.Done()

	//listen for TCP conns on the specified port
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	//sets the last the func will do before exiting, close the connections
	defer ln.Close()

	fmt.Println("Awaiting TCP connection from Client")

	//for each connection, accept it, and then place it in its own goroutine
	for {
		connection, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection Error: ", err)
			break
		}
		fmt.Println("Accepted connection", connection)
		go HandleConn(connection)

	}
	mainwg.Done()

}

//Function to be ran in its own goroutine that handles the data sent from the connected client
//The function receives input from the client, and capitlizes it, then sends it back.
func HandleConn(conn net.Conn) {
	for {

		fmt.Println("awaiting message")

		//read the string from the connection
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("received msg: ", message)

		//have the message capitilized
		newmessage := strings.ToUpper(message)

		// send new string back to client
		conn.Write([]byte(newmessage + "\n"))
	}
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
