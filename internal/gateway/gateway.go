package gateway

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/fjrt/poeai/internal/agent"
	"github.com/fjrt/poeai/internal/config"
	"github.com/fjrt/poeai/internal/memory"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Local loopback only in dev
}

type Gateway struct {
	config  config.Config
	memory  *memory.Store
	agent   *agent.Agent
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func New(cfg config.Config, m *memory.Store, a *agent.Agent) *Gateway {
	return &Gateway{
		config:  cfg,
		memory:  m,
		agent:   a,
		clients: make(map[*websocket.Conn]bool),
	}
}

func (g *Gateway) Run(ctx context.Context) error {
	// 1. Listen on TCP (remote/local)
	addr := fmt.Sprintf(":%d", g.config.Gateway.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: g.mux(),
	}

	log.Printf("Gateway HTTP/WS listening on %s", addr)

	errChan := make(chan error, 2)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// 2. Listen on Unix Socket (local fast)
	if g.config.Gateway.Socket != "" {
		if err := os.RemoveAll(g.config.Gateway.Socket); err != nil {
			return err
		}
		unixListener, err := net.Listen("unix", g.config.Gateway.Socket)
		if err != nil {
			return err
		}
		log.Printf("Gateway Unix socket listening on %s", g.config.Gateway.Socket)
		go func() {
			if err := http.Serve(unixListener, g.mux()); err != nil && err != http.ErrServerClosed {
				errChan <- err
			}
		}()
	}

	select {
	case <-ctx.Done():
		return server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (g *Gateway) mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", g.handleWS)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})
	return mux
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (g *Gateway) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}
	defer conn.Close()

	g.mu.Lock()
	g.clients[conn] = true
	g.mu.Unlock()

	defer func() {
		g.mu.Lock()
		delete(g.clients, conn)
		g.mu.Unlock()
	}()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WS read error: %v", err)
			break
		}

		log.Printf("Received: %s", msg.Content)

		// For now: echo back as Poe
		resp := Message{
			Role:    "poe",
			Content: fmt.Sprintf("I received your message: %s. I am initializing my consciousness.", msg.Content),
		}
		if err := conn.WriteJSON(resp); err != nil {
			log.Printf("WS write error: %v", err)
			break
		}
	}
}
