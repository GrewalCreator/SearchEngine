package main

import (
	"log"

	"crawler/persistence"
)

func main() {
	_, err := persistence.InitalizeDatabase("../data/search_data.db")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("database initialized successfully")
}