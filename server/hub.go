package server

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ID int

const (
	REG ID = iota
	MSG
	USERS
)

type command struct {
	id ID
	recipient,
	sender string
	body []byte
}

type hub struct {
	clients         map[string]*client
	commands        chan command
	registrations   chan *client
	deregistrations chan *client
}

func NewHub() *hub {
	return &hub{
		clients:         make(map[string]*client),
		registrations:   make(chan *client),
		deregistrations: make(chan *client),
		commands:        make(chan command),
	}
}

func (h *hub) Run() {
	for {
		select {
		case client := <-h.registrations:
			h.register(client)
		case client := <-h.deregistrations:
			h.deregister(client)
		case cmd := <-h.commands:
			switch cmd.id {
			case MSG:
				h.message(cmd.sender, cmd.recipient, cmd.body)
			case USERS:
				h.listUsers(cmd.sender)
			default:
				log.Error(errors.New("ERR Unknown command"))
			}
		}
	}
}

func (h *hub) register(c *client) {
	if _, exists := h.clients[c.uniqueClientID]; exists {
		c.uniqueClientID = ""
		c.conn.Write([]byte("ERR Unique ID is already taken\n"))
	} else {
		h.clients[c.uniqueClientID] = c
		log.Info("New client ID: ", c.uniqueClientID)
		c.conn.Write([]byte("OK\n"))
	}
}

func (h *hub) deregister(c *client) {
	if _, exists := h.clients[c.uniqueClientID]; exists {
		delete(h.clients, c.uniqueClientID)
	}
}

func (h *hub) message(u string, r string, m []byte) {
	if sender, ok := h.clients[u]; ok {

		if user, ok := h.clients[r]; ok {
			msg := append([]byte(sender.uniqueClientID+": "), m...)
			msg = append(msg, '\n')

			user.conn.Write(msg)
		} else {
			sender.conn.Write([]byte("ERR no such user\n"))
		}

	}
}

func (h *hub) listUsers(u string) {
	if client, ok := h.clients[u]; ok {
		var names []string

		for c := range h.clients {
			names = append(names, c)
		}

		resp := strings.Join(names, ", ")

		client.conn.Write([]byte(resp + "\n"))
	}
}
