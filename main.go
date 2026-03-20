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
		c.HTML(http.StatusOK, "index.html", gin.H{
			"roomList": app.RoomRepo.GetRoomList(),
		})
	})

	r.GET("/room/:id", func(c *gin.Context) {
		roomId, err := uuid.Parse(c.Param("id"))

		if err != nil {
			pageNotFound(c)
			return
		}

		room := app.RoomRepo.GetById(roomId)

		if room == nil {
			pageNotFound(c)
			return
		}

		c.HTML(http.StatusOK, "room.html", gin.H{
			"name": room.Name,
		})
	})

	r.POST("/create-room", createRoomHandler)
	r.POST("/delete-room", deleteRoomHandler)

	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)
	r.GET("/me", meHandler)
	r.GET("/room/:id/info", roomInfoHandler)

	r.GET("/room/:id/ws", websocketConnectHandler)

	fmt.Println("Starting...")

	app = newApp()

	log.Fatal(r.Run(":8080"))
}

func pageNotFound(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", nil)
}
