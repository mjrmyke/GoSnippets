//Myke Walker
//8-13-2017
//TCP Server Program to receive a Client Connection
//Client sends a string, Server ToUppers the string, then returns
//
// go run server.go -port=3333
//
// port flag is which port the Server will listen on
// run program with -h for help with flags and to see defaults

package main

import (
	"bufio"
	"flag"
	"fmt"
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

}

//Function to be ran in its own goroutine that handles the data sent from the connected client
//The function receives input from the client, and capitalizes it, then sends it back.
func HandleConn(conn net.Conn) {
	r := bufio.NewReader(conn)
	for {

		fmt.Println("awaiting message")

		//read the string from the connection
		message, err := r.ReadString('\n')
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
