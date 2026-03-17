package main

import "github.com/google/uuid"

type LoginRequest struct {
	Name string `json:"name"`
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
	WebsocketEventTypeDisconnect    = WebsocketEventType("disconnect")

	WebsocketMessageTypeSend   = WebsocketMessageType("message-send")
	WebsocketMessageTypeUpdate = WebsocketMessageType("message-update")
)

type WebsocketEvent struct {
	User    User
	Type    WebsocketEventType
	Message Message
}
