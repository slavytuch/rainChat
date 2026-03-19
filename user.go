package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math/rand"
)

type User struct {
	id     uuid.UUID
	Name   string `json:"name"`
	Color  string `json:"color"`
	sendCh chan Message
}

func NewUser(name string) User {
	return User{
		id:     uuid.New(),
		Name:   name,
		Color:  fmt.Sprintf("#%.2x%.2x%.2x", rand.Intn(256), rand.Intn(256), rand.Intn(256)),
		sendCh: make(chan Message),
	}
}

func (u *User) createClient(conn *websocket.Conn) Client {
	return Client{
		Id:     uuid.New(),
		User:   u,
		conn:   conn,
		readCh: make(chan WebsocketEvent),
		doneCh: make(chan struct{}),
	}
}
