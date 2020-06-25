package main

import (
	"log"
)

func main() {
	server := &Server{Port: "6000"}

	log.Printf("serving on port %s", server.Port)
	server.ListenAndServe()
}
