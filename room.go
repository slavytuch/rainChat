package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log/slog"
	"time"
)

type Room struct {
	Id             uuid.UUID
	ConnectionList map[User]*websocket.Conn
	MessageList    []Message
	ReadChannel    chan WebsocketEvent
	WriteChannel   chan Message
}

type Message struct {
	Id        uuid.UUID `json:"id"`
	Author    string    `json:"author"`
	Color     string    `json:"color"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	Updated   bool      `json:"updated"`
}

func userMessage(user User, message string) Message {
	return Message{
		Id:        uuid.New(),
		Author:    user.Name,
		Color:     user.Color,
		Text:      message,
		CreatedAt: time.Now(),
		Updated:   false,
	}
}

func systemMessage(message string) Message {
	return userMessage(User{
		ID:    uuid.New(),
		Name:  "System",
		Color: "red",
	}, message)
}

func (r *Room) Connect(user User, conn *websocket.Conn) error {

	if _, ok := r.ConnectionList[user]; ok {
		return errors.New("user is already connected")
	}

	for u, _ := range r.ConnectionList {
		u.SendCh <- systemMessage(fmt.Sprintf("User %s connected", user.Name))
	}

	r.ConnectionList[user] = conn

	for _, m := range r.MessageList {
		conn.WriteJSON(m)
	}

	return nil
}

func (r *Room) Disconnect(user User) {
	conn, ok := r.ConnectionList[user]

	if !ok {
		return
	}

	conn.WriteMessage(websocket.TextMessage, []byte("Connection closed by server"))
	conn.Close()

	delete(r.ConnectionList, user)
}

func NewRoom() Room {
	return Room{
		Id:             uuid.New(),
		ConnectionList: make(map[User]*websocket.Conn),
		ReadChannel:    make(chan WebsocketEvent),
	}
}

func (r *Room) Serve(doneCh chan bool) {
	for {
		select {
		case <-doneCh:
			fmt.Println("Close room received")
			close(r.ReadChannel)
			return
		case we := <-r.ReadChannel:
			slog.Info("Received event", "event", we)
			switch we.Type {
			case WebsocketEventTypeMessageSend:
				r.MessageList = append(r.MessageList, we.Message)

				for u, _ := range r.ConnectionList {
					u.SendCh <- we.Message
				}
			case WebsocketEventTypeDisconnect:
				r.Disconnect(we.User)
				for u, _ := range r.ConnectionList {
					u.SendCh <- systemMessage(fmt.Sprintf("User %s has disconnected", we.User.Name))
				}
			case WebsocketEventTypeMessageUpdate:
				found := false
				for midx, m := range r.MessageList {
					if m.Id != we.Message.Id {
						continue
					}

					r.MessageList[midx] = we.Message
					r.MessageList[midx].Updated = true
					found = true

					for _, c := range r.ConnectionList {
						c.WriteJSON(r.MessageList[midx])
					}
				}

				if !found {
					r.ConnectionList[we.User].WriteJSON(map[string]any{
						"error": "message not found",
					})
				}
			default:
				slog.Error("Unknown event type")
			}
		}
	}
}
