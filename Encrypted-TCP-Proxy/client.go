//Myke Walker
//8-13-2017
//TCP Client Program to request use of a Proxy
//Encrypted Packet sent to the proxy determines where the final destination should be
//Receives Encrypted packet from proxy that contains the response of the final destination server.

//
// go run client.go -key=1234567890123456 -dport=3333 -pport=2222
//
// key flag is for encryption key (16,24,32 bytes)
// pport flag is the port for the proxy Server
// dport flag is the port of the destination Server
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
	"os"
)

//Struct to hold simulated packet data
//all fields are []byte so they may be encrypted
type Packet struct {
	To      []byte `json: "To"`
	From    []byte `json: "From"`
	Message []byte `json: "Message"`
}

//global secret key for encryption. Length of Key determines  16 = AES128, 24 = AES192, 32 = AES256
var AESkey = flag.String("key", "1234567890123456", "Encryption Key - 16, 24, or 32 byes")

func main() {

	//set flags, and default values
	proxyport := flag.String("pport", "2222", "port of proxy to connect to")
	destport := flag.String("dport", "3333", "port to have proxy connect to")

	//parse flags
	flag.Parse()

	//ensure the key is the correct size
	switch len(*AESkey) {
	case 16:
		break
	case 24:
		break
	case 32:
		break
	default:
		fmt.Println("length of key is: ", len(*AESkey))
		fmt.Println("AESKey must be 16, 24, or 32 bytes")
		os.Exit(1)
	}

	//Dial the connection to the proxy
	conn, err := net.Dial("tcp", "127.0.0.1:"+*proxyport)
	if err != nil {
		log.Fatal(err)
	}

	//Encrypt the Proxy Request Port
	OutgoingPacket := Packet{
		To: encrypt([]byte(*destport)),
	}

	//while loop to handle communication with the proxy
	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		//write the message to the packet and encrypt it
		OutgoingPacket.Message = encrypt([]byte(text))

		//marshal the struct to json
		tmpdata, err := json.Marshal(OutgoingPacket)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("sent")
		// send to socket
		_, err = conn.Write(tmpdata)
		if err != nil {
			log.Fatal(err)
		}

		// add json decoder to connection
		d := json.NewDecoder(conn)

		var packet Packet

		fmt.Println("awaiting packet")

		//decode the packet from connection
		err = d.Decode(&packet)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("received encrypted message")

		decryptedmsg := decrypt(packet.Message)

		fmt.Println("decrypted Message from server: " + string(decryptedmsg))
		fmt.Println("\n")
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
