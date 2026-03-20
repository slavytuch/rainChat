package main

import "github.com/google/uuid"

type LoginRequest struct {
	Name string `form:"name" binding:"required"`
}

type RegisterRequest struct {
	Name string `form:"name" binding:"required"`
}

type CreateRoomRequest struct {
	Name string `form:"name" binding:"required"`
}

type DeleteRoomRequest struct {
	Id uuid.UUID `form:"id" binding:"required"`
}
