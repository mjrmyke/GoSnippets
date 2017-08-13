# Encrypted TCP Proxy
Assortment of programs and tools in Golang
Proxy Program to act as an intermediary between a server and client
Client sends an encrypted request to the TCP connection of the proxy, and the proxy
decrypts it, and forwards it to its desired destination. The proxy also receives unencrypted data from the destination
and encrypts it, then ships it back to the original client

To use, start the proxy and server before the client. IE:

go run proxy.go

go run server.go

go run client.go

