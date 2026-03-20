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

	var userList []map[string]string

	for _, conn := range room.ConnectionList {
		userList = append(userList, map[string]string{
			"id":    conn.Id.String(),
			"name":  conn.User.Name,
			"color": conn.User.Color,
		})
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
	token, err := c.Cookie("user-token")

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error reading cookie: " + err.Error(),
		})
		return
	}

	userId, err := uuid.Parse(token)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid token: " + err.Error(),
		})
		return
	}

	user := app.UserRepo.GetById(userId)

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
