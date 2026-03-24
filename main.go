package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
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

	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)

	r.POST("/create-room", createRoomHandler)
	r.POST("/delete-room", deleteRoomHandler)

	r.GET("/room/:id", roomHandler)
	r.GET("/room/:id/info", roomInfoHandler)

	r.GET("/reflect", func(c *gin.Context) {
		c.HTML(http.StatusOK, "reflect.html", nil)
	})

	r.POST("/reflect-connect", reflectConnectHandler)

	{
		authGroup := r.Group("/")
		authGroup.Use(authRequired())
		authGroup.GET("/me", meHandler)
		authGroup.GET("/room/:id/ws", websocketConnectHandler)
	}

	fmt.Println("Starting...")

	app = newApp()

	log.Fatal(r.Run(":8080"))
}

func pageNotFound(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", nil)
}
