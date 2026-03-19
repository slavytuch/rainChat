package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math/rand"
)

type User struct {
	ID     uuid.UUID
	Name   string
	Color  string
	SendCh chan Message
}

func NewUser(name string) User {
	return User{
		ID:     uuid.New(),
		Name:   name,
		Color:  fmt.Sprintf("#%.2x%.2x%.2x", rand.Intn(256), rand.Intn(256), rand.Intn(256)),
		SendCh: make(chan Message),
	}
}

func (u *User) createClient(conn *websocket.Conn) Client {
	return Client{
		Id:     uuid.New(),
		User:   u,
		conn:   conn,
		ReadCh: make(chan WebsocketEvent),
	}
}
