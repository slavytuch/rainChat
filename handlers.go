package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"rainChat/internal/chat"
)

func badHeaderResp(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
}

func loginHandler(c *gin.Context) {
	var req LoginRequest

	err := c.BindJSON(&req)

	if err != nil {
		badHeaderResp(c, "invalid reading request body: "+err.Error())
		return
	}

	if req.Name == "" {
		badHeaderResp(c, "name is empty")
		return
	}

	user := app.UserRepo.GetByName(req.Name)

	if user == nil {
		badHeaderResp(c, "not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": user.Id,
		"name":  user.Name,
		"color": user.Color,
	})
}

func registerHandler(c *gin.Context) {
	var req RegisterRequest

	err := c.BindJSON(&req)

	if err != nil {
		badHeaderResp(c, "invalid reading request body: "+err.Error())
		return
	}

	if req.Name == "" {
		badHeaderResp(c, "name is empty")
		return
	}

	newUser := chat.NewUser(req.Name)
	err = app.UserRepo.PushUser(&newUser)

	if err != nil {
		badHeaderResp(c, "error creating user: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": newUser.Id,
		"name":  newUser.Name,
		"color": newUser.Color,
	})
}

func roomHandler(c *gin.Context) {
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
}

func roomInfoHandler(c *gin.Context) {
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

	var userList []*chat.User

	for _, conn := range room.ConnectionList {
		userList = append(userList, conn.User)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       room.Id.String(),
		"name":     room.Name,
		"userList": userList,
	})
}

func websocketConnectHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("Error from upgrade: " + err.Error())
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

	connUser := c.MustGet("user").(*chat.User)

	client := connUser.CreateClient(conn)
	go client.Reader(room.BroadcastingChannel)
	go client.Writer(room.BroadcastingChannel)

	err = room.Connect(&client)

	if err != nil {
		writeWebsocketError(conn, "Error connecting to room: "+err.Error())
		return
	}
}

func createRoomHandler(c *gin.Context) {
	var req CreateRoomRequest

	err := c.BindJSON(&req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})

		return
	}

	room := chat.NewRoom(req.Name)

	doneCh := app.RoomRepo.Push(&room)

	go room.Serve(doneCh)

	c.JSON(http.StatusCreated, gin.H{
		"id":   room.Id,
		"name": room.Name,
		"link": fmt.Sprintf("/room/%s", room.Id),
	})
}

func deleteRoomHandler(c *gin.Context) {
	var req DeleteRoomRequest

	err := c.BindJSON(&req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "internal error",
		})

		return
	}

	app.RoomRepo.DeleteById(req.Id)
}

func writeWebsocketError(conn *websocket.Conn, msg string) {
	conn.WriteJSON(gin.H{"error": msg})
	conn.Close()
}

func meHandler(c *gin.Context) {
	user := c.MustGet("user").(*chat.User)

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
