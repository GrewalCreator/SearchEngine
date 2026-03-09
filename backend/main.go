package main

import (
	"log"

	"crawler/server"
)

func main() {
	srv, err := server.NewServer()
	if err != nil {
		log.Fatal(err)
	}

	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}