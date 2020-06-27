package main

import (
	"github.com/maulanakmal/ums/rpc"
	"log"
)

func main() {
	server := &rpc.Server{Port: "6000"}

	log.Printf("serving on port %s", server.Port)
	server.ListenAndServe()
}
