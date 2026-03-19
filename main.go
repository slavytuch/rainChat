package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
var app *App

func main() {
	r := gin.Default()

	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("pages/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)
	r.GET("/room/:id/user-list", roomUserListHandler)

	r.GET("/room/:id/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Error from upgrade: " + err.Error())
			return
		}

		roomId, err := uuid.Parse(c.Param("id"))

		if err != nil {
			writeWebsocketError(conn, "Unknown room")
			return
		}

		room := app.RoomRepo.GetById(roomId)

		if room == nil {
			writeWebsocketError(conn, "Unknown room")
			return
		}

		userId := c.Query("token")

		if userId == "" {
			writeWebsocketError(conn, "Token is invalid")
			return
		}

		userUUID, err := uuid.Parse(userId)

		if err != nil {
			writeWebsocketError(conn, "Token is invalid")
			return
		}

		connUser := app.UserRepo.GetById(userUUID)

		if connUser == nil {
			writeWebsocketError(conn, "Token is invalid")
			return
		}

		client := connUser.createClient(conn)
		go client.reader(room.BroadcastingChannel)
		go client.writer(room.BroadcastingChannel)

		err = room.Connect(&client)

		if err != nil {
			writeWebsocketError(conn, "Error connecting to room: "+err.Error())
			return
		}
	})

	fmt.Println("Starting...")

	app = newApp()

	log.Fatal(r.Run(":8080"))
}

func writeWebsocketError(conn *websocket.Conn, msg string) {
	conn.WriteJSON(gin.H{"error": msg})
	conn.Close()
}
