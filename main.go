package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var userRepo UserRepo
var room = NewRoom()

func main() {
	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, "pages/index.html")
	})

	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)

	mux.HandleFunc("/ws", func(rw http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			log.Println("Error from upgrade: " + err.Error())
			return
		}

		userId := r.URL.Query().Get("token")

		if userId == "" {
			conn.WriteMessage(websocket.TextMessage, []byte("UserId header is not provided"))
			conn.Close()
			return
		}

		userUUID, err := uuid.Parse(userId)

		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("UserId is invalid"))
			conn.Close()
			return
		}

		connUser := userRepo.GetById(userUUID)

		if connUser == nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Client not found"))
			conn.Close()
			return
		}

		client := connUser.createClient(conn)
		go client.reader(room.BroadcastingChannel)
		go client.writer(room.BroadcastingChannel)

		err = room.Connect(&client)

		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Error connecting to room: "+err.Error()))
			conn.Close()
			return
		}
	})

	fmt.Println("Starting...")
	go room.Serve(make(chan bool))
	log.Fatal(http.ListenAndServe(":8080", mux))
}
