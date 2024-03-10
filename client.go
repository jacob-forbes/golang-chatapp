package main

import (
	"bytes"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	name    string
	socket  *websocket.Conn
	receive chan []byte
	room    *room
}

var (
	readSep  = []byte(": ")
	writeSep = []byte(" ")
)

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, wsMsg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		msg := bytes.Join([][]byte{[]byte(c.name), wsMsg}, readSep)
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.receive {
		time := time.Now().UTC().Format("2006-01-02 15:04:05")
		wsMsg := bytes.Join([][]byte{[]byte(time), msg}, writeSep)
		err := c.socket.WriteMessage(websocket.TextMessage, wsMsg)
		if err != nil {
			return
		}
	}
}
