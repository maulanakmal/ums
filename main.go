package main

import (
	"encoding/gob"
	"log"
	"net"
)

type LoginRequest struct {
	Username string
	Password string
}

type RegisterRequest struct {
	Username string
	Nickname string
	Password string
}

type ChangeNicknameRequest struct {
	Username string
	Nickname string
}

func main() {
	server := &Server{Port: "6000"}
	go server.ListenAndServe()

	conn, err := net.Dial("tcp", "localhost:6000")
	if err != nil {
		panic(err.Error())
	}
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)
	request := Request{Name: "changeNickname", Args: []string{"mol2", "molana"}}
	var response Response
	encoder.Encode(request)
	decoder.Decode(&response)
	log.Printf("response %q", response)

}
