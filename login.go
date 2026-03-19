package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
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
		"token": user.id,
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

	newUser := NewUser(req.Name)
	err = app.UserRepo.PushUser(&newUser)

	if err != nil {
		badHeaderResp(c, "error creating user: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": newUser.id,
		"name":  newUser.Name,
		"color": newUser.Color,
	})
}
