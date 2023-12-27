package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/HyggeHalcyon/go-broadcast-server/server"
)

func main() {
	// parsing arguments
	var addr string
	flag.StringVar(&addr, "l", "127.0.0.1:8080", "address:port to listen on")
	flag.Parse()

	// initialization
	fmt.Println("[!] Starting Server...")
	listen, err := net.Listen("tcp", addr)
	defer listen.Close()
	if err != nil {
		log.Fatalf("error listening: %s", err.Error())
	}
	fmt.Printf("[!] listening on %s\n", listen.Addr().String())

	// pipelines for communicating between broadcasting thread and clients
	pipe := make(chan server.Payload, 1)
	quit := make(chan server.Client, 1)

	// start the broadcasting thread
	clientsPool := server.NewThreadSafeClients(pipe, quit)
	clientsPool.Dispatch()   // thread for broadcasting messages
	clientsPool.Disconnect() // thread for detecting when a client disconnects

	// listen for incoming clients
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("failed accepting: ", err.Error())
			conn.Close()
			continue
		}

		// initialize the client object with the connection and pipelines
		client := server.NewClient(conn, pipe, quit)

		// add the connection to the pool to trace clients and broadcast messages
		clientsPool.Push(client)

		// start the client thread to listen for input
		client.Start()
	}
}
