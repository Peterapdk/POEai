package main

import (
	"log"
	"net/url"

	"github.com/fjrt/poeai/internal/tui"
	"github.com/gorilla/websocket"
)

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:7331", Path: "/ws"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	if err := tui.Run(conn); err != nil {
		log.Fatalf("tui: %v", err)
	}
}
