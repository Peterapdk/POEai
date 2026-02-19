package memory_test

import (
	"context"
	"testing"

	"github.com/fjrt/poeai/internal/memory"
)

func TestStore_WriteAndSearch(t *testing.T) {
	store, err := memory.Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	id, err := store.Write(ctx, memory.Memory{
		Type:    memory.TypeEpisodic,
		Content: "fjrt fixed the ESPHome thermostat node",
		Source:  "conversation",
	})
	if err != nil || id == "" {
		t.Fatalf("Write() error = %v, id = %q", err, id)
	}

	results, err := store.Search(ctx, "thermostat", 5)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) == 0 {
		t.Error("Search() returned no results, want at least 1")
	}
}

func TestStore_Facts(t *testing.T) {
	store, _ := memory.Open(":memory:")
	defer store.Close()
	ctx := context.Background()

	if err := store.SetFact(ctx, "homelab.ha-server.os", "Debian 12", 1.0); err != nil {
		t.Fatalf("SetFact() error = %v", err)
	}
	val, ok, err := store.GetFact(ctx, "homelab.ha-server.os")
	if err != nil || !ok || val != "Debian 12" {
		t.Errorf("GetFact() = %q, %v, %v", val, ok, err)
	}
}
