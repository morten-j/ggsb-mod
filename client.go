package main

import (
	"bufio"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"strings"
)

type client struct {
	conn     net.Conn
	nick     string
	commands chan<- command
	private  *rsa.PrivateKey
	public   rsa.PublicKey
}

func (c *client) readInput() {
	for {
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			return
		}

		msg = strings.Trim(msg, "\r\n")
		args := strings.Split(msg, " ")
		cmd := strings.TrimSpace(args[0])

		switch cmd {
		case "/nick":
			c.commands <- command{
				id:     CMD_NICK,
				client: c,
				args:   args,
			}
		case "/clients":
			c.commands <- command{
				id:     CMD_ROOMS,
				client: c,
				args:   args,
			}
		case "/msg":
			c.commands <- command{
				id:     CMD_MSG,
				client: c,
				args:   args,
			}
		case "/quit":
			c.commands <- command{
				id:     CMD_QUIT,
				client: c,
				args:   args,
			}
		default:
			c.err(fmt.Errorf("Unknown command: %s", cmd))
		}
	}
}

func (c *client) err(err error) {
	c.conn.Write([]byte("ERR: " + err.Error() + "\n"))
}

func (c *client) msg(x *client, msg string) {

	//Check if contacting other client (Decrypt or not)
	if c.private != x.private {
		dMsg := decrypt(msg, *x.private)
		_, e := x.conn.Write([]byte("=> " + dMsg + "\n"))
		if e != nil {
			log.Fatalln("unable to write over client connection")
		}

	} else {
		_, e := x.conn.Write([]byte("=> " + msg + "\n"))
		if e != nil {
			log.Fatalln("unable to write over client connection")
		}
	}

}
