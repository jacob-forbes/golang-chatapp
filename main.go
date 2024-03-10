package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}
var rooms = make(map[string]*room)
var nameGenerator = namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	roomName := strings.TrimPrefix(r.URL.Path, "/ws/")
	room, exists := rooms[roomName]
	if !exists {
		log.Printf("Requested room %v does not exist yet, creating new room.", roomName)
		room = newRoom()
		go room.run()
		rooms[roomName] = room
	}

	log.Printf("Upgrading connection...")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade err:", err)
		return
	}

	log.Printf("Creating client...")
	client := &client{
		name:    nameGenerator.Generate(),
		socket:  conn,
		receive: make(chan []byte, messageBufferSize),
		room:    room,
	}

	log.Printf("Joining room...")
	room.join <- client
	room.forward <- []byte(client.name + " has joined the room!")
	defer func() {
		room.leave <- client
		room.forward <- []byte(client.name + " has left the room.")
	}()
	go client.write()
	client.read()
}

func main() {
	http.HandleFunc("/ws/", websocketHandler)

	log.Println("Listening for connections!")
	err := http.ListenAndServe(":5005", nil)
	if err != nil {
		log.Println("http server err: ", err)
		return
	}
}
