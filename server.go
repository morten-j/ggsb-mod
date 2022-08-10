package main

import (
	"bufio"
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

//TODO Skal tilføje brugernavn (måske kodeord også???)
func (s *server) newClient(conn net.Conn) {
	// generate RSA keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalln(err)
	}

	username := promptConnection(conn, "What username would you like?")

	c := &client{
		conn:     conn,
		nick:     username,
		commands: s.commands,
		private:  privateKey,
		public:   privateKey.PublicKey,
	}

	s.clients[c.nick] = c

	log.Printf("New client connected: %s", conn.RemoteAddr().String())

	c.msg(c, "Welcome to the server!"+"\n")

	c.readInput()
}

func promptConnection(connection net.Conn, prompt string) string {
	_, e := connection.Write([]byte(prompt + "\n"))
	if e != nil {
		log.Fatalln("unable to write over client connection")
	}

	msg, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		log.Fatalln("unable to read commandline")
	}

	return msg
}

//TODO Check if name is valid
func (s *server) nick(c *client, args []string) {
	if len(args) > 2 {
		c.msg(c, "There can not be spaces in usernames!"+"\n")
		return
	}

	c.nick = args[1]

	s.clients[c.nick] = c

	c.msg(c, fmt.Sprintf("Name changed to %s", c.nick))
}

//TODO Fix så det virker med clients
func (s *server) listClients(c *client, args []string) {
	var clients []string
	for name := range s.clients {
		clients = append(clients, name)
	}

	c.msg(c, fmt.Sprintf("Rooms available: %s", strings.Join(clients, ", ")))
}

//TODO Security check
func (s *server) msg(c *client, args []string) {
	//Check if client exist on server and use it for check
	r, ok := s.clients[args[1]]
	if ok {
		//Format the message (Concat 2. argument and everything after it)
		msg := strings.Join(args[2:], " ")
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
