package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log/slog"
	"math/rand"
	"time"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
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

func (u User) reader(conn *websocket.Conn, receiveCh chan<- WebsocketEvent) {
	defer func() {
		receiveCh <- WebsocketEvent{
			User: u,
			Type: WebsocketEventTypeDisconnect,
		}
		conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	var msg WebsocketMessage
	for {
		err := conn.ReadJSON(&msg)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Error reading message", "user", u, "error", err)
			}
			break
		}

		slog.Info("Received message", "message", msg)

		switch msg.Type {
		case WebsocketMessageTypeSend:
			receiveCh <- WebsocketEvent{
				User:    u,
				Type:    WebsocketEventTypeMessageSend,
				Message: userMessage(u, msg.Text),
			}
		case WebsocketMessageTypeUpdate:
			um := userMessage(u, msg.Text)
			um.Id = msg.MessageId
			receiveCh <- WebsocketEvent{
				User:    u,
				Type:    WebsocketEventTypeMessageUpdate,
				Message: um,
			}
		default:
			slog.Error("Unknown message type")
		}
	}
}

func (u *User) writer(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case msg, ok := <-u.SendCh:
			slog.Info("Sending message", "message", msg)
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte("Read channel closed"))
				return
			}

			conn.WriteJSON(msg)
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type Repo struct {
	userList []User
}

func (r *Repo) PushUser(user User) error {
	for _, u := range r.userList {
		if u.Name == user.Name {
			return fmt.Errorf("user with name %s already exists", user.Name)
		}
	}

	r.userList = append(r.userList, user)
	return nil
}

func (r *Repo) GetById(id uuid.UUID) *User {
	for _, u := range r.userList {
		if u.ID == id {
			return &u
		}
	}

	return nil
}

func (r *Repo) DeleteUser(user User) {
	var result []User
	for _, u := range r.userList {
		if u.ID != user.ID {
			result = append(result, u)
		}
	}

	r.userList = result
}
