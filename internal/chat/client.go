package chat

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log/slog"
	"sync/atomic"
	"time"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

type Client struct {
	Id     uuid.UUID `json:"Id"`
	User   *User     `json:"user"`
	conn   *websocket.Conn
	readCh chan WebsocketEvent
	closed atomic.Bool
	doneCh chan struct{}
}

func (c *Client) Close() {
	if c.closed.Load() {
		return
	}

	c.closed.Store(true)

	slog.Info("closing Client via external call", "Client", c)
	c.conn.Close()
	c.doneCh <- struct{}{}
}

func (c *Client) Reader(roomCh chan<- WebsocketEvent) {
	defer func() {
		roomCh <- WebsocketEvent{
			Type:   WebsocketEventTypeDisconnect,
			Client: c,
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	var msg WebsocketMessage

	for {
		err := c.conn.ReadJSON(&msg)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("Error reading message", "Client", c, "error", err)
			} else {
				slog.Info("Closing ws connection expectedly")
			}
			break
		}

		slog.Info("Received message", "message", msg)

		switch msg.Type {
		case WebsocketMessageTypeSend:
			roomCh <- WebsocketEvent{
				Client:  c,
				Type:    WebsocketEventTypeMessageSend,
				Message: userMessage(c.User, msg.Text),
			}
		case WebsocketMessageTypeUpdate:
			um := userMessage(c.User, msg.Text)
			um.Id = msg.MessageId
			roomCh <- WebsocketEvent{
				Client:  c,
				Type:    WebsocketEventTypeMessageUpdate,
				Message: um,
			}
		default:
			slog.Error("Unknown message type")
		}
	}
}

func (c *Client) Writer(roomCh chan<- WebsocketEvent) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		roomCh <- WebsocketEvent{
			Type:   WebsocketEventTypeDisconnect,
			Client: c,
		}
		ticker.Stop()
	}()

	for {
		select {
		case <-c.doneCh:
			slog.Info("done received", "client", c)
			return
		case msg, ok := <-c.readCh:
			slog.Info("Sending message", "message", msg, "Client", c)
			if !ok {
				slog.Info("Read channel closed", "Client", c)
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.conn.WriteJSON(msg)
			if err != nil {
				slog.Error("Error sending message", "error", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Ping error", "error", err)
				return
			}
		}
	}
}
