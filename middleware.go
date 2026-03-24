package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("user-token")

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Error reading access token: " + err.Error(),
			})
			return
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Auth required",
			})
			return
		}

		userId, err := uuid.Parse(token)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Invalid access token",
			})
			return
		}

		user := app.UserRepo.GetById(userId)

		if user == nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "User not found",
			})
			return
		}

		c.Set("user", user)

		c.Next()
	}
}
