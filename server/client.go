package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"

	log "github.com/sirupsen/logrus"
)

type client struct {
	conn           net.Conn
	outbound       chan<- command
	register       chan<- *client
	deregister     chan<- *client
	uniqueClientID string
}

func NewClient(con net.Conn, out chan<- command, reg chan<- *client, dreg chan<- *client) *client {
	return &client{
		conn:       con,
		outbound:   out,
		register:   reg,
		deregister: dreg,
	}
}

func (c *client) Read() {

	// Say hi to the new client
	// c.conn.Write([]byte("WELCOME\n"))
	log.Info("New client connection")

	defer c.conn.Close()
	reader := bufio.NewReader(c.conn)

	for {
		data, err := reader.ReadBytes('\n')
		switch err {
		case io.EOF:
			// Connection has been closed by client
			log.Info(c.uniqueClientID, " left")
			c.deregister <- c
			log.Info("Client deregistered")
			return
		case nil:
			c.handle(data)
		default:
			log.Error(err)
			return
		}
	}
}

func (c *client) handle(data []byte) {

	cmd := bytes.ToUpper(bytes.TrimSpace(bytes.Split(data, []byte(" "))[0]))
	payload := bytes.TrimSpace(bytes.TrimPrefix(data, cmd))

	switch string(cmd) {
	case "REG":
		if err := c.reg(payload); err != nil {
			c.err(err)
		}
	case "MSG":
		if err := c.msg(payload); err != nil {
			c.err(err)
		}
	case "USERS":
		c.usrs()
	default:
		c.err(fmt.Errorf("Unknown command"))
	}
}

func (c *client) reg(payload []byte) error {

	if len(payload) == 0 {
		log.Error("A user name must start with @")
		return fmt.Errorf("A user name must start with @")
	}

	if payload[0] != '@' {
		log.Error("A user name must start with @")
		return fmt.Errorf("A user name must start with @")
	}

	if len(payload) < 4 {
		log.Error("A user name must have a min len of 3 chars")
		return fmt.Errorf("A user name must have a min len of 3 chars")
	}

	c.uniqueClientID = string(bytes.TrimSpace(payload))
	c.register <- c

	return nil
}

func (c *client) msg(payload []byte) error {

	if payload[0] != '@' {
		log.Error("Recipient must start with '@'; a channel ID must start with '#'")
		return fmt.Errorf("Recipient must start with '@'; a channel ID must start with '#'")
	}

	if len(payload) < 4 {
		log.Error("Recipient must have a min len of 3 chars")
		return fmt.Errorf("Recipient must have a min len of 3 chars")
	}

	recipient := bytes.TrimSpace(bytes.Split(payload, []byte(" "))[0])
	msgBdy := bytes.TrimSpace(bytes.TrimPrefix(payload, recipient))

	c.outbound <- command{
		recipient: string(recipient),
		sender:    c.uniqueClientID,
		body:      msgBdy,
		id:        MSG,
	}

	return nil
}

func (c *client) usrs() {
	c.outbound <- command{
		sender: c.uniqueClientID,
		id:     USERS,
	}
}

func (c *client) err(e error) {
	c.conn.Write([]byte("ERR " + e.Error() + "\n"))
}
