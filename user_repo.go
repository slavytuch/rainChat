package main

import (
	"fmt"
	"github.com/google/uuid"
)

type UserRepo struct {
	userList []*User
}

func (r *UserRepo) PushUser(user *User) error {
	for _, u := range r.userList {
		if u.Name == user.Name {
			return fmt.Errorf("user with name %s already exists", user.Name)
		}
	}

	r.userList = append(r.userList, user)
	return nil
}

func (r *UserRepo) GetById(id uuid.UUID) *User {
	for i, u := range r.userList {
		if u.id == id {
			return r.userList[i]
		}
	}

	return nil
}

func (r *UserRepo) GetByName(name string) *User {
	for i, u := range r.userList {
		if u.Name == name {
			return r.userList[i]
		}
	}

	return nil
}

func (r *UserRepo) DeleteUser(user User) {
	var result []*User
	for _, u := range r.userList {
		if u.id != user.id {
			result = append(result, u)
		}
	}

	r.userList = result
}
