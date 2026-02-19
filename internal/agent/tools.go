package agent

import (
	"context"
	"fmt"

	"github.com/fjrt/poeai/internal/memory"
)

func (a *Agent) registerCoreTools() {
	a.RegisterTool("memory_write", a.toolMemoryWrite)
	a.RegisterTool("memory_search", a.toolMemorySearch)
}

func (a *Agent) toolMemoryWrite(ctx context.Context, params map[string]interface{}) (string, error) {
	content, ok := params["content"].(string)
	if !ok {
		return "", fmt.Errorf("missing content")
	}
	mTypeStr, _ := params["type"].(string)
	mType := memory.TypeEpisodic
	if mTypeStr != "" {
		mType = memory.MemoryType(mTypeStr)
	}

	id, err := a.memory.Write(ctx, memory.Memory{
		Type:    mType,
		Content: content,
		Source:  "agent",
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Memory stored with ID: %s", id), nil
}

func (a *Agent) toolMemorySearch(ctx context.Context, params map[string]interface{}) (string, error) {
	query, ok := params["query"].(string)
	if !ok {
		return "", fmt.Errorf("missing query")
	}
	limit := 5
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	results, err := a.memory.Search(ctx, query, limit)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "No relevant memories found.", nil
	}

	out := "Relevant memories:\n"
	for _, r := range results {
		out += fmt.Sprintf("- [%s] %s\n", r.CreatedAt.Format("2006-01-02"), r.Content)
	}
	return out, nil
}
