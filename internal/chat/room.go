package chat

import (
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type Room struct {
	Id                  uuid.UUID           `json:"id"`
	Name                string              `json:"name"`
	ConnectionList      []*Client           `json:"-"`
	MessageList         []Message           `json:"-"`
	BroadcastingChannel chan WebsocketEvent `json:"-"`
}

type Message struct {
	Id        uuid.UUID `json:"Id"`
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

func (r *Room) Connect(client *Client) error {
	slog.Info("Connecting Client", "Client", client)

	for _, c := range r.ConnectionList {
		c.readCh <- WebsocketEvent{
			Client: client,
			Type:   WebsocketEventTypeConnect,
		}
	}

	for _, m := range r.MessageList {
		client.readCh <- WebsocketEvent{
			Type:    WebsocketEventTypeMessageSend,
			Message: m,
		}
	}

	r.ConnectionList = append(r.ConnectionList, client)

	return nil
}

func (r *Room) deleteClient(client *Client) {
	for i, c := range r.ConnectionList {
		if c.Id == client.Id {
			r.ConnectionList[i] = r.ConnectionList[len(r.ConnectionList)-1]
			r.ConnectionList = r.ConnectionList[:len(r.ConnectionList)-1]
			break
		}
	}
}

func NewRoom(name string) Room {
	return Room{
		Id:                  uuid.New(),
		Name:                name,
		BroadcastingChannel: make(chan WebsocketEvent),
	}
}

func (r *Room) Serve(doneCh chan struct{}) {
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
					c.readCh <- WebsocketEvent{
						Client:  we.Client,
						Type:    WebsocketEventTypeMessageSend,
						Message: we.Message,
					}
				}
			case WebsocketEventTypeDisconnect:
				r.deleteClient(we.Client)
				we.Client.Close()
				for _, c := range r.ConnectionList {
					c.readCh <- WebsocketEvent{
						Type:   WebsocketEventTypeDisconnect,
						Client: we.Client,
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
						c.readCh <- WebsocketEvent{
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
