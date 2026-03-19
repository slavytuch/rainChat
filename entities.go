package main

import "github.com/google/uuid"

type LoginRequest struct {
	Name string `form:"name" binding:"required"`
}

type RegisterRequest struct {
	Name string `form:"name" binding:"required"`
}

type WebsocketMessage struct {
	Type      WebsocketMessageType `json:"type"`
	Text      string               `json:"text"`
	MessageId uuid.UUID            `json:"messageId"`
}

type WebsocketEventType string
type WebsocketMessageType string

const (
	WebsocketEventTypeMessageSend   = WebsocketEventType("message-send")
	WebsocketEventTypeMessageUpdate = WebsocketEventType("message-update")
	WebsocketEventTypeConnect       = WebsocketEventType("connect")
	WebsocketEventTypeDisconnect    = WebsocketEventType("disconnect")

	WebsocketMessageTypeSend   = WebsocketMessageType("message-send")
	WebsocketMessageTypeUpdate = WebsocketMessageType("message-update")
)

type WebsocketEvent struct {
	Client  *Client            `json:"client"`
	Type    WebsocketEventType `json:"type"`
	Message Message            `json:"message"`
}
