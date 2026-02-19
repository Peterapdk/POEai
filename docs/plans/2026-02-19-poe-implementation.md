# Poe AI Sidekick Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build Poe — a Go monorepo with poe-gateway daemon, poe TUI client, and poe-node Android companion.

**Architecture:** Three binaries sharing internal packages. Gateway runs as systemd service. TUI connects via Unix socket. Android node (Termux) pushes sensors to gateway via HTTP.

**Tech Stack:** Go 1.22+, Bubbletea, Lipgloss, Glamour, SQLite + sqlite-vec, golang.org/x/crypto/ssh, gorilla/websocket, BurntSushi/toml

---

### Task 1: Go Module + Project Skeleton

**Files:**
- Create: `go.mod`
- Create: `go.sum` (auto-generated)
- Create: `Makefile`
- Create: `cmd/poe-gateway/main.go`
- Create: `cmd/poe/main.go`
- Create: `cmd/poe-node/main.go`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `~/.poe/config.toml` (sample)

**Step 1: Init go module**
```bash
cd /home/fjrt/POEai
go mod init github.com/fjrt/poeai
```
Expected: `go.mod` created with `module github.com/fjrt/poeai` and `go 1.22`

**Step 2: Write failing test for config loading**
```go
// internal/config/config_test.go
package config_test

import (
    "os"
    "testing"
    "github.com/fjrt/poeai/internal/config"
)

func TestLoadConfig_Defaults(t *testing.T) {
    cfg, err := config.Load("")
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }
    if cfg.Gateway.Port != 7331 {
        t.Errorf("default port = %d, want 7331", cfg.Gateway.Port)
    }
}

func TestLoadConfig_FromFile(t *testing.T) {
    f, _ := os.CreateTemp("", "poe-config-*.toml")
    f.WriteString("[gateway]\nport = 9999\n")
    f.Close()
    defer os.Remove(f.Name())

    cfg, err := config.Load(f.Name())
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }
    if cfg.Gateway.Port != 9999 {
        t.Errorf("port = %d, want 9999", cfg.Gateway.Port)
    }
}
```

**Step 3: Run test — expect FAIL (package not found)**
```bash
go test ./internal/config/... -v
```
Expected: compile error

**Step 4: Implement config package**
```go
// internal/config/config.go
package config

import (
    "os"
    "path/filepath"
    "github.com/BurntSushi/toml"
)

type Config struct {
    LLM     LLMConfig     `toml:"llm"`
    Gateway GatewayConfig `toml:"gateway"`
    Memory  MemoryConfig  `toml:"memory"`
    Nodes   map[string]NodeConfig `toml:"nodes"`
}

type LLMConfig struct {
    Provider string `toml:"provider"`
    Model    string `toml:"model"`
    APIKey   string `toml:"api_key"`
}

type GatewayConfig struct {
    Socket string `toml:"socket"`
    Port   int    `toml:"port"`
}

type MemoryConfig struct {
    DBPath         string `toml:"db_path"`
    EmbeddingModel string `toml:"embedding_model"`
}

type NodeConfig struct {
    Host string `toml:"host"`
    User string `toml:"user"`
    Key  string `toml:"key"`
}

func defaults() Config {
    home, _ := os.UserHomeDir()
    return Config{
        LLM: LLMConfig{
            Provider: "anthropic",
            Model:    "claude-opus-4-6",
        },
        Gateway: GatewayConfig{
            Socket: filepath.Join(home, ".poe", "poe.sock"),
            Port:   7331,
        },
        Memory: MemoryConfig{
            DBPath:         filepath.Join(home, ".poe", "poe.db"),
            EmbeddingModel: "ollama/nomic-embed-text",
        },
    }
}

// Load loads config from path (empty = defaults only).
func Load(path string) (Config, error) {
    cfg := defaults()
    if path == "" {
        return cfg, nil
    }
    _, err := toml.DecodeFile(path, &cfg)
    return cfg, err
}
```

**Step 5: Add dependency**
```bash
go get github.com/BurntSushi/toml
```

**Step 6: Run test — expect PASS**
```bash
go test ./internal/config/... -v -race
```
Expected: PASS, no race conditions

**Step 7: Create stub entry points**
```go
// cmd/poe-gateway/main.go
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
    fmt.Printf("Poe Gateway starting on port %d\n", cfg.Gateway.Port)
}
```
(Same minimal stub pattern for `cmd/poe/main.go` and `cmd/poe-node/main.go`)

**Step 8: Create Makefile**
```makefile
.PHONY: build test lint

build:
	go build ./cmd/...

test:
	go test ./... -race -v

lint:
	golangci-lint run ./...
```

**Step 9: Build all**
```bash
go build ./cmd/...
```
Expected: three binaries compile with no errors

**Step 10: Commit**
```bash
git add .
git commit -m "feat: go module init, config package, binary stubs"
```

---

### Task 2: Memory Package (SQLite + sqlite-vec)

**Files:**
- Create: `internal/memory/memory.go`
- Create: `internal/memory/memory_test.go`
- Create: `internal/memory/schema.sql`

**Step 1: Write failing tests**
```go
// internal/memory/memory_test.go
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
```

**Step 2: Run — expect FAIL**
```bash
go test ./internal/memory/... -v
```

**Step 3: Implement memory package**

Schema SQL (`internal/memory/schema.sql`):
```sql
CREATE TABLE IF NOT EXISTS memories (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL,
    content     TEXT NOT NULL,
    source      TEXT NOT NULL,
    importance  REAL NOT NULL DEFAULT 0.5,
    created_at  INTEGER NOT NULL,
    accessed_at INTEGER NOT NULL,
    metadata    TEXT NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS facts (
    key         TEXT PRIMARY KEY,
    value       TEXT NOT NULL,
    confidence  REAL NOT NULL DEFAULT 1.0,
    updated_at  INTEGER NOT NULL
);
```

Core types and Open/Close/Write/Search/SetFact/GetFact methods in `internal/memory/memory.go` using `github.com/mattn/go-sqlite3`. For search without vector embeddings in v1: use SQLite FTS5 full-text search (sqlite-vec added in Task 5 when embeddings are wired up).

**Step 4: Add dependencies**
```bash
go get github.com/mattn/go-sqlite3
```

**Step 5: Run tests — expect PASS**
```bash
go test ./internal/memory/... -v -race
```

**Step 6: Commit**
```bash
git add .
git commit -m "feat: memory package with SQLite FTS5 backend"
```

---

### Task 3: SSH Node Controller

**Files:**
- Create: `internal/ssh/client.go`
- Create: `internal/ssh/client_test.go`

**Step 1: Write failing tests (uses mock SSH server)**
Test that `Exec(ctx, cmd)` returns stdout/stderr/exit-code correctly.

**Step 2: Run — expect FAIL**

**Step 3: Implement**
```go
// internal/ssh/client.go
package ssh

import (
    "bytes"
    "context"
    "fmt"
    "os"
    "golang.org/x/crypto/ssh"
)

type Client struct {
    host string
    user string
    key  string
}

type Result struct {
    Stdout   string
    Stderr   string
    ExitCode int
}

func New(host, user, keyPath string) *Client {
    return &Client{host: host, user: user, key: keyPath}
}

func (c *Client) Exec(ctx context.Context, cmd string) (Result, error) {
    key, err := os.ReadFile(c.key)
    if err != nil {
        return Result{}, fmt.Errorf("read key: %w", err)
    }
    signer, err := ssh.ParsePrivateKey(key)
    if err != nil {
        return Result{}, fmt.Errorf("parse key: %w", err)
    }
    cfg := &ssh.ClientConfig{
        User:            c.user,
        Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
        HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: known_hosts
    }
    conn, err := ssh.Dial("tcp", c.host+":22", cfg)
    if err != nil {
        return Result{}, fmt.Errorf("dial %s: %w", c.host, err)
    }
    defer conn.Close()

    sess, err := conn.NewSession()
    if err != nil {
        return Result{}, fmt.Errorf("session: %w", err)
    }
    defer sess.Close()

    var stdout, stderr bytes.Buffer
    sess.Stdout = &stdout
    sess.Stderr = &stderr

    exitCode := 0
    if err := sess.Run(cmd); err != nil {
        if exitErr, ok := err.(*ssh.ExitError); ok {
            exitCode = exitErr.ExitStatus()
        } else {
            return Result{}, fmt.Errorf("run: %w", err)
        }
    }
    return Result{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}, nil
}
```

**Step 4: Add dependency**
```bash
go get golang.org/x/crypto/ssh
```

**Step 5: Run tests — expect PASS**
```bash
go test ./internal/ssh/... -v -race
```

**Step 6: Commit**
```bash
git add .
git commit -m "feat: SSH node controller"
```

---

### Task 4: poe-node Android Companion HTTP Server

**Files:**
- Create: `internal/node/server.go`
- Create: `internal/node/server_test.go`
- Create: `cmd/poe-node/main.go` (replace stub)

**Step 1: Write failing tests**
Test that GET `/status` returns JSON with required fields (battery, location, activity, wifi).
Test that POST `/notify` returns 200.

**Step 2: Run — expect FAIL**

**Step 3: Implement**
```go
// internal/node/server.go
package node

import (
    "encoding/json"
    "net/http"
)

type Status struct {
    Battery  BatteryInfo  `json:"battery"`
    Location LocationInfo `json:"location"`
    Activity string       `json:"activity"`
    Wifi     WifiInfo     `json:"wifi"`
}
type BatteryInfo  struct { Level int `json:"level"`; Charging bool `json:"charging"` }
type LocationInfo struct { Lat float64 `json:"lat"`; Lon float64 `json:"lon"`; Accuracy float64 `json:"accuracy"` }
type WifiInfo     struct { SSID string `json:"ssid"`; Connected bool `json:"connected"` }

type Server struct {
    status Status
    mux    *http.ServeMux
}

func New() *Server {
    s := &Server{mux: http.NewServeMux()}
    s.mux.HandleFunc("GET /status", s.handleStatus)
    s.mux.HandleFunc("POST /notify", s.handleNotify)
    s.mux.HandleFunc("GET /pair", s.handlePair)
    return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(s.status)
}

func (s *Server) handleNotify(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePair(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
}

func (s *Server) ListenAndServe(addr string) error {
    return http.ListenAndServe(addr, s)
}
```

**Step 4: Run tests — expect PASS**
```bash
go test ./internal/node/... -v -race
```

**Step 5: Wire into cmd/poe-node/main.go**

**Step 6: Build**
```bash
go build ./cmd/poe-node/
```

**Step 7: Commit**
```bash
git add .
git commit -m "feat: poe-node HTTP server for Android companion"
```

---

### Task 5: Gateway Agent Loop + Tool Dispatch

**Files:**
- Create: `internal/agent/agent.go`
- Create: `internal/agent/tools.go`
- Create: `internal/agent/agent_test.go`

**Goal:** Agent struct with `Run(ctx)` loop, tool registry, and basic tool implementations (ssh_exec, memory_write, memory_search, notify_phone, soul_update).

**Step 1: Write failing tests**
Test that agent correctly dispatches to tool handlers when given a mock LLM response containing a tool call.

**Step 2: Run — expect FAIL**

**Step 3: Implement agent with tool dispatch loop**

Key design: Agent has a `map[string]ToolFunc` registry. LLM responses with tool calls are decoded and dispatched. Each tool returns a string result that is fed back into the next LLM turn.

```go
// internal/agent/agent.go
package agent

import "context"

type ToolFunc func(ctx context.Context, params map[string]any) (string, error)

type Agent struct {
    tools  map[string]ToolFunc
    memory MemoryStore  // interface
    llm    LLMClient    // interface
}

func New(mem MemoryStore, llm LLMClient) *Agent {
    a := &Agent{
        tools:  make(map[string]ToolFunc),
        memory: mem,
        llm:    llm,
    }
    a.registerCoreTools()
    return a
}

func (a *Agent) RegisterTool(name string, fn ToolFunc) {
    a.tools[name] = fn
}

// Run is the autonomous background loop.
func (a *Agent) Run(ctx context.Context) error {
    // tick-based loop with configurable interval
    // each tick: gather state → retrieve memories → call LLM → dispatch tools
    return nil // TODO: implement in Task 6
}
```

**Step 4: Run tests — expect PASS**
```bash
go test ./internal/agent/... -v -race
```

**Step 5: Commit**
```bash
git add .
git commit -m "feat: agent loop with tool dispatch registry"
```

---

### Task 6: poe-gateway Daemon (WebSocket + Unix Socket)

**Files:**
- Create: `internal/gateway/gateway.go`  
- Create: `internal/gateway/gateway_test.go`
- Create: `cmd/poe-gateway/main.go` (replace stub)
- Create: `scripts/poe-gateway.service` (systemd unit)

**Goal:** Gateway starts memory store, agent loop, and WebSocket server. TUI connects via Unix socket.

**Step 1: Write failing test**
Test that gateway starts, accepts a WebSocket connection, handles a chat message, and returns a response.

**Step 2: Run — expect FAIL**

**Step 3: Implement gateway**
```go
// internal/gateway/gateway.go  
package gateway

import (
    "context"
    "net"
    "net/http"
    "github.com/fjrt/poeai/internal/agent"
    "github.com/fjrt/poeai/internal/config"
    "github.com/fjrt/poeai/internal/memory"
    "github.com/gorilla/websocket"
)
```

Systemd unit (`scripts/poe-gateway.service`):
```ini
[Unit]
Description=Poe AI Gateway
After=network.target

[Service]
Type=simple
ExecStart=%h/.local/bin/poe-gateway
Restart=always
RestartSec=5
Environment=HOME=%h

[Install]
WantedBy=default.target
```

**Step 4: Add dependency**
```bash
go get github.com/gorilla/websocket
```

**Step 5: Build + smoke test**
```bash
go build ./cmd/poe-gateway/
./poe-gateway &
sleep 1
curl -s http://localhost:7331/health
kill %1
```

**Step 6: Commit**
```bash
git add .
git commit -m "feat: poe-gateway WebSocket daemon + systemd unit"
```

---

### Task 7: poe TUI Client (Bubbletea)

**Files:**
- Create: `internal/tui/model.go`
- Create: `internal/tui/model_test.go`
- Create: `internal/tui/styles.go`
- Create: `cmd/poe/main.go` (replace stub)

**Goal:** Full-screen Bubbletea TUI. Chat view with Poe's responses rendered in Glamour markdown. Status bar showing node + memory count. Dark gothic aesthetic.

**Step 1: Write failing test**
Test that initial model renders without panicking, status bar contains expected text.

**Step 2: Run — expect FAIL**

**Step 3: Add Charm dependencies**
```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/glamour
go get github.com/charmbracelet/bubbles
```

**Step 4: Implement styles**
```go
// internal/tui/styles.go
package tui

import "github.com/charmbracelet/lipgloss"

var (
    colorBg       = lipgloss.Color("#0d0d0d")
    colorAccent   = lipgloss.Color("#8b5cf6")  // violet — Poe's signature
    colorDim      = lipgloss.Color("#4b4b4b")
    colorText     = lipgloss.Color("#e2e2e2")
    colorPoeText  = lipgloss.Color("#c4b5fd")
    colorUserText = lipgloss.Color("#6ee7b7")

    styleHeader = lipgloss.NewStyle().
        Foreground(colorAccent).
        Bold(true).
        Padding(0, 1)

    styleStatusBar = lipgloss.NewStyle().
        Foreground(colorDim).
        Padding(0, 1)

    styleInput = lipgloss.NewStyle().
        Foreground(colorText).
        BorderStyle(lipgloss.NormalBorder()).
        BorderForeground(colorAccent).
        Padding(0, 1)

    stylePoeMsg  = lipgloss.NewStyle().Foreground(colorPoeText)
    styleUserMsg = lipgloss.NewStyle().Foreground(colorUserText)
)
```

**Step 5: Implement Bubbletea model** (model.go)
- `Init()` → connect to gateway Unix socket
- `Update()` → handle keypresses, WebSocket messages
- `View()` → render header + chat history + input box

**Step 6: Run tests — expect PASS**
```bash
go test ./internal/tui/... -v -race
```

**Step 7: Build + launch**
```bash
go build ./cmd/poe/
./poe
```
Verify: TUI renders, violet accent, gothic dark theme.

**Step 8: Commit**
```bash
git add .
git commit -m "feat: Bubbletea TUI client with dark violet theme"
```

---

### Task 8: SOUL.md + Personality (AGENTS.md)

**Files:**
- Create: `internal/soul/soul.go`
- Create: `internal/soul/soul_test.go`
- Create: `~/.poe/AGENTS.md` (written on first run)

**Goal:** Soul package reads/writes SOUL.md. Gateway writes initial AGENTS.md prompt on first run. Memory consolidation goroutine appends learned patterns to SOUL.md.

**Step 1: Write failing tests**
Test Load/Update/GetPrompt on SOUL.md.

**Step 2: Implement**

**Step 3: Run — expect PASS**
```bash
go test ./internal/soul/... -v -race
```

**Step 4: Wire soul prompt into agent context on each LLM call**

**Step 5: Commit**
```bash
git add .
git commit -m "feat: SOUL.md procedural memory + AGENTS.md personality prompt"
```

---

### Task 9: Integration + README

**Files:**
- Create: `README.md`
- Modify: `Makefile` (add install target)

**Goal:** Wire all packages together end-to-end. Install target copies binaries to `~/.local/bin`. README covers setup.

**Step 1: Integration smoke test**
Start gateway, connect TUI, send a message, verify response arrives.

**Step 2: Install target**
```makefile
install:
	go build -o ~/.local/bin/poe-gateway ./cmd/poe-gateway/
	go build -o ~/.local/bin/poe ./cmd/poe/
	go build -o ~/.local/bin/poe-node ./cmd/poe-node/
	systemctl --user daemon-reload
	systemctl --user enable poe-gateway
```

**Step 3: Write README.md**

**Step 4: Final build + test**
```bash
make test
make build
```

**Step 5: Commit**
```bash
git add .
git commit -m "feat: integration, install target, README"
```
