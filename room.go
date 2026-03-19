package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
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

func NewRoom() Room {
	return Room{
		Id:                  uuid.MustParse("3e813ad4-b88d-4af1-b55c-43f8552ba32e"),
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

func roomUserListHandler(c *gin.Context) {
	roomId, err := uuid.Parse(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	room := app.RoomRepo.GetById(roomId)

	if room == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var userList []map[string]string

	for _, c := range room.ConnectionList {
		userList = append(userList, map[string]string{
			"id":    c.Id.String(),
			"name":  c.User.Name,
			"color": c.User.Color,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"userList": userList,
	})
}
