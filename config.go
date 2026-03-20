package main

import "rainChat/internal/chat"

type App struct {
	UserRepo chat.UserRepo
	RoomRepo chat.RoomRepo
}

func newApp() *App {
	return &App{
		RoomRepo: chat.RoomRepo{RoomList: make(map[*chat.Room]chan struct{})},
	}
}
