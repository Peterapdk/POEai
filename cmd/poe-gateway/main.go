package main

import (
	"fmt"
	"log"

	"github.com/fjrt/poeai/internal/config"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	fmt.Printf("Poe Gateway — port %d, db %s\n", cfg.Gateway.Port, cfg.Memory.DBPath)
}
