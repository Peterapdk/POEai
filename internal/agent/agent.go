package agent

import (
	"context"
	"fmt"

	"github.com/fjrt/poeai/internal/memory"
)

type ToolFunc func(ctx context.Context, params map[string]interface{}) (string, error)

type Agent struct {
	tools  map[string]ToolFunc
	memory *memory.Store
}

func New(m *memory.Store) *Agent {
	a := &Agent{
		tools:  make(map[string]ToolFunc),
		memory: m,
	}
	a.registerCoreTools()
	return a
}

func (a *Agent) RegisterTool(name string, fn ToolFunc) {
	a.tools[name] = fn
}

func (a *Agent) Dispatch(ctx context.Context, name string, params map[string]interface{}) (string, error) {
	fn, ok := a.tools[name]
	if !ok {
		return "", fmt.Errorf("tool %q not found", name)
	}
	return fn(ctx, params)
}
