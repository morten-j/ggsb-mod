package main

import (
	"log"
	"net"
)

//TODO * Database for login
// * Save messages encrypted on PC
// * Save private key local on PC

func main() {
	s := newServer()
	go s.run()

	listener, err := net.Listen("tcp", ":8888")

	if err != nil {
		log.Fatalf("Could not start server: %s", err.Error())
	}

	defer listener.Close()
	log.Printf("Started server on :8888")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("unable to accept connection: %s", err.Error())
			continue
		}

		go s.newClient(conn)
	}
}
