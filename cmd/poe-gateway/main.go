package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fjrt/poeai/internal/agent"
	"github.com/fjrt/poeai/internal/config"
	"github.com/fjrt/poeai/internal/gateway"
	"github.com/fjrt/poeai/internal/memory"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Ensure .poe directory exists
	if err := os.MkdirAll(os.ExpandEnv("$HOME/.poe"), 0755); err != nil {
		log.Fatalf("mkdir: %v", err)
	}

	mem, err := memory.Open(cfg.Memory.DBPath)
	if err != nil {
		log.Fatalf("memory: %v", err)
	}
	defer mem.Close()

	age := agent.New(mem)
	gtw := gateway.New(cfg, mem, age)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Poe AI Sidekick — starting daemon")
	if err := gtw.Run(ctx); err != nil {
		log.Fatalf("gateway: %v", err)
	}
}
