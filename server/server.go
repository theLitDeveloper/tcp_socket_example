package server

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

func Run(port int16) {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// Start the hub
	hub := NewHub()
	go hub.Run()

	log.Info("Server is listening on port ", port)

	// Wait and listen for connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error(err.(net.Error))
		}

		// Create a new client, when a new connection is accepted
		clnt := NewClient(
			conn,
			hub.commands,
			hub.registrations,
			hub.deregistrations,
		)

		// Read messages sent by client
		go clnt.Read()
	}
}
