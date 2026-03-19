package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func badHeaderResp(rw http.ResponseWriter, msg string) {
	rw.WriteHeader(http.StatusBadRequest)
	respBody, _ := json.Marshal(map[string]any{
		"error": msg,
	})
	rw.Write(respBody)
}

func loginHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req LoginRequest

	body, err := io.ReadAll(r.Body)

	if err != nil {
		badHeaderResp(rw, "invalid reading request body: "+err.Error())
		return
	}

	err = json.Unmarshal(body, &req)

	if err != nil {
		badHeaderResp(rw, "invalid request body: "+err.Error())
		return
	}

	if req.Name == "" {
		badHeaderResp(rw, "name is empty")
		return
	}

	user := userRepo.GetByName(req.Name)

	if user == nil {
		badHeaderResp(rw, "not found")
		return
	}
	respBody, _ := json.Marshal(map[string]any{
		"token": user.ID,
	})
	rw.Write(respBody)
}

func registerHandler(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req LoginRequest

	body, err := io.ReadAll(r.Body)

	if err != nil {
		badHeaderResp(rw, "invalid reading request body: "+err.Error())
		return
	}

	err = json.Unmarshal(body, &req)

	if err != nil {
		badHeaderResp(rw, "invalid request body: "+err.Error())
		return
	}

	if req.Name == "" {
		badHeaderResp(rw, "name is empty")
		return
	}

	newUser := NewUser(req.Name)
	err = userRepo.PushUser(newUser)

	if err != nil {
		badHeaderResp(rw, "error creating user: "+err.Error())
		return
	}

	rw.WriteHeader(http.StatusCreated)

	respBody, _ := json.Marshal(map[string]any{
		"token": newUser.ID,
	})
	rw.Write(respBody)
}
