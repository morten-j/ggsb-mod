package main

type commandID int

const (
	CMD_NICK = iota //Increment constant
	CMD_JOIN
	CMD_ROOMS
	CMD_MSG
	CMD_QUIT
)

type command struct {
	id     commandID
	client *client
	args   []string
}
