package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"strings"
)

type server struct {
	clients  map[string]*client
	commands chan command
}

func newServer() *server {
	return &server{
		clients:  make(map[string]*client),
		commands: make(chan command),
	}
}

func (s *server) run() {
	for cmd := range s.commands {
		switch cmd.id {
		case CMD_NICK:
			s.nick(cmd.client, cmd.args)
		case CMD_ROOMS:
			s.listClients(cmd.client, cmd.args)
		case CMD_MSG:
			s.msg(cmd.client, cmd.args)
		case CMD_QUIT:
			s.quit(cmd.client, cmd.args)
		}
	}
}

func (s *server) newClient(conn net.Conn) {
	// generate RSA keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalln(err)
	}

	c := &client{
		conn:     conn,
		nick:     "anon",
		commands: s.commands,
		private:  privateKey,
		public:   privateKey.PublicKey,
	}

	log.Printf("New client connected: %s", conn.RemoteAddr().String())

	c.readInput()
}

func (s *server) nick(c *client, args []string) {
	c.nick = args[1]

	s.clients[c.nick] = c

	c.msg(c, fmt.Sprintf("Name changed to %s", c.nick))
}

func (s *server) listClients(c *client, args []string) {
	var clients []string
	for name := range s.clients {
		clients = append(clients, name)
	}

	c.msg(c, fmt.Sprintf("Rooms available: %s", strings.Join(clients, ", ")))
}

func (s *server) msg(c *client, args []string) {
	//Check if client exist on server and use it for check
	r, ok := s.clients[args[1]]
	if ok {
		//Format the message
		msg := strings.Join(args[1:], " ")
		msg = c.nick + " : " + msg

		//Get public key of the reciever of the message
		publicKey := r.public

		//Encrypt
		eMsg := encrypt(msg, publicKey)

		//Send message
		c.msg(r, eMsg)
	}
}

func (s *server) quit(c *client, args []string) {
	log.Printf("Client has disconnected: %s", c.conn.RemoteAddr().String())

	//Remove from client list
	_, ok := s.clients[c.nick]
	if ok {
		delete(s.clients, c.nick)
	}

	c.msg(c, "Bye bb!")
	c.conn.Close()
}
