package main

import (
	"log"

	"github.com/fjrt/poeai/internal/node"
)

func main() {
	srv := node.New()
	log.Println("Poe Node starting on :7332")
	if err := srv.ListenAndServe(":7332"); err != nil {
		log.Fatalf("node server: %v", err)
	}
}
