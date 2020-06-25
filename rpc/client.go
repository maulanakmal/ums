package rpc

import (
	"encoding/gob"
	"errors"
	"net"
)

type Client struct {
	Addr string
}

func (client *Client) Call(request Request) (*Response, error) {
	if client.Addr == "" {
		return nil, errors.New("client addr is not set")
	}

	conn, err := net.Dial("tcp", client.Addr)
	if err != nil {
		panic(err.Error())
	}

	encoder := gob.NewEncoder(conn)
	encoder.Encode(request)
	decoder := gob.NewDecoder(conn)

	var response Response
	decoder.Decode(&response)

	return &response, nil
}
