package main

import "github.com/google/uuid"

type RoomRepo struct {
	RoomList []*Room
}

func (rr *RoomRepo) Push(r *Room) {
	rr.RoomList = append(rr.RoomList, r)
}

func (rr *RoomRepo) GetById(id uuid.UUID) *Room {
	for i, r := range rr.RoomList {
		if r.Id == id {
			return rr.RoomList[i]
		}
	}

	return nil
}
