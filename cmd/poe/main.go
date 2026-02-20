package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/fjrt/poeai/internal/onboarding"
	"github.com/fjrt/poeai/internal/tui"
	"github.com/gorilla/websocket"
)

func main() {
	home, _ := os.UserHomeDir()
	configPath := filepath.Join(home, ".poe", "config.toml")

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "configure":
			_, err := onboarding.Configure()
			if err != nil {
				log.Fatalf("Configure failed: %v", err)
			}
			return
		case "onboarding":
			_, err := onboarding.Onboard()
			if err != nil {
				log.Fatalf("Onboarding failed: %v", err)
			}
			fmt.Println("Configuration saved. Please start the gateway with: poe-gateway")
			return
		}
	}

	// Try to dial gateway
	u := url.URL{Scheme: "ws", Host: "localhost:7331", Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		// Gateway is down. Check if we need onboarding.
		if _, statErr := os.Stat(configPath); os.IsNotExist(statErr) {
			_, err := onboarding.Onboard()
			if err != nil {
				log.Fatalf("Onboarding failed: %v", err)
			}
			fmt.Println("Configuration saved. Please start the gateway with: poe-gateway")
			return
		}
		log.Fatalf("Poe Gateway is not running. Please start it with: poe-gateway (Error: %v)", err)
	}
	defer conn.Close()

	if err := tui.Run(conn); err != nil {
		log.Fatalf("tui: %v", err)
	}
}
