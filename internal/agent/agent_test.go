package agent_test

import (
	"context"
	"testing"

	"github.com/fjrt/poeai/internal/agent"
	"github.com/fjrt/poeai/internal/memory"
)

func TestAgent_ToolDispatch(t *testing.T) {
	mem, _ := memory.Open(":memory:")
	defer mem.Close()

	a := agent.New(mem)
	ctx := context.Background()

	// Test memory_write via Dispatch
	params := map[string]interface{}{
		"content": "test memory",
		"type":    "episodic",
	}
	res, err := a.Dispatch(ctx, "memory_write", params)
	if err != nil {
		t.Fatalf("Dispatch(memory_write) error = %v", err)
	}
	if res == "" {
		t.Error("Dispatch(memory_write) returned empty result")
	}

	// Test memory_search via Dispatch
	searchParams := map[string]interface{}{
		"query": "test",
	}
	searchRes, err := a.Dispatch(ctx, "memory_search", searchParams)
	if err != nil {
		t.Fatalf("Dispatch(memory_search) error = %v", err)
	}
	if searchRes == "No relevant memories found." {
		t.Error("Search should have found the memory we just wrote")
	}
}
