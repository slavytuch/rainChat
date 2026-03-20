package chat

import (
	"github.com/google/uuid"
)

type RoomRepo struct {
	RoomList map[*Room]chan struct{}
}

func (rr *RoomRepo) Push(r *Room) (doneCh chan struct{}) {
	doneCh = make(chan struct{})
	rr.RoomList[r] = doneCh
	return doneCh
}

func (rr *RoomRepo) GetById(id uuid.UUID) *Room {
	for r := range rr.RoomList {
		if r.Id == id {
			return r
		}
	}

	return nil
}

func (rr *RoomRepo) DeleteById(id uuid.UUID) {
	for r, d := range rr.RoomList {
		if r.Id == id {
			d <- struct{}{}
			delete(rr.RoomList, r)
			break
		}
	}
}

func (rr *RoomRepo) GetRoomList() (res []*Room) {
	for r := range rr.RoomList {
		res = append(res, r)
	}

	return
}
