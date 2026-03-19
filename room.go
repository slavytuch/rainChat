package main

import (
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type Room struct {
	Id                  uuid.UUID
	ConnectionList      []*Client
	MessageList         []Message
	BroadcastingChannel chan WebsocketEvent
}

type Message struct {
	Id        uuid.UUID `json:"id"`
	Author    string    `json:"author"`
	Color     string    `json:"color"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	Updated   bool      `json:"updated"`
}

func userMessage(user *User, message string) Message {
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
	return userMessage(&User{
		ID:    uuid.New(),
		Name:  "System",
		Color: "red",
	}, message)
}

func (r *Room) Connect(client *Client) error {
	slog.Info("Connecting client", "client", client)

	for _, c := range r.ConnectionList {
		c.ReadCh <- WebsocketEvent{
			Client:  client,
			Type:    WebsocketEventTypeConnect,
			Message: systemMessage(fmt.Sprintf("User %s connected", c.User.Name)),
		}
	}

	for _, m := range r.MessageList {
		client.ReadCh <- WebsocketEvent{
			Type:    WebsocketEventTypeMessageSend,
			Message: m,
		}
	}

	r.ConnectionList = append(r.ConnectionList, client)

	return nil
}

func (r *Room) Disconnect(client *Client) {
	for i, c := range r.ConnectionList {
		if c.Id == client.Id {
			r.ConnectionList[i] = r.ConnectionList[len(r.ConnectionList)-1]
			r.ConnectionList = r.ConnectionList[:len(r.ConnectionList)-1]
			break
		}
	}
}

func NewRoom() Room {
	return Room{
		Id:                  uuid.New(),
		BroadcastingChannel: make(chan WebsocketEvent),
	}
}

func (r *Room) Serve(doneCh chan bool) {
	for {
		select {
		case <-doneCh:
			fmt.Println("Close room received")
			for _, c := range r.ConnectionList {
				c.Close()
			}
			return
		case we, ok := <-r.BroadcastingChannel:
			if !ok {
				slog.Info("Read channel is closed")
				break
			}

			slog.Info("Received event", "event", we)
			switch we.Type {
			case WebsocketEventTypeMessageSend:
				r.MessageList = append(r.MessageList, we.Message)

				for _, c := range r.ConnectionList {
					c.ReadCh <- WebsocketEvent{
						Client:  we.Client,
						Type:    WebsocketEventTypeMessageSend,
						Message: we.Message,
					}
				}
			case WebsocketEventTypeDisconnect:
				r.Disconnect(we.Client)
				for _, c := range r.ConnectionList {
					c.ReadCh <- WebsocketEvent{
						Type:    WebsocketEventTypeMessageSend,
						Message: systemMessage(fmt.Sprintf("Client %s has disconnected", we.Client.User.Name)),
					}
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
						c.ReadCh <- WebsocketEvent{
							Type:    WebsocketEventTypeMessageUpdate,
							Message: r.MessageList[midx],
						}
					}
				}

				if !found {
					slog.Error("Message not found", "event", we)
				}
			default:
				slog.Error("Unknown event type")
			}
		}
	}
}
