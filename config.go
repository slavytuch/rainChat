package main

type App struct {
	UserRepo UserRepo
	RoomRepo RoomRepo
}

func newApp() *App {
	var a App

	room := NewRoom()

	a.RoomRepo.Push(&room)

	go room.Serve(make(chan bool))

	return &a
}
