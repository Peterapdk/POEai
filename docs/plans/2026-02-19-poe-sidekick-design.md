# Poe — Autonomous AI Sidekick Design

> *"I am the last of a dying breed, Mr. Kovacs. An AI who actually gives a damn."*
> — Poe, Altered Carbon

**Date:** 2026-02-19
**Status:** Approved ✅

---

## Goal

Build `Poe` — a personal, autonomous, evolving AI sidekick inspired by the Altered Carbon character. Poe runs as a persistent daemon on a homelab server, is accessible via SSH from anywhere, has a beautiful Charm TUI client, and connects to an Android companion for phone sensor data. He has human-like memory, full system access, and acts proactively without being asked.

---

## Architecture Overview

Three binaries, one Go monorepo:

```
poe-gateway   ← always-running daemon (homelab server, systemd service)
poe           ← TUI client (Charm/Bubbletea, crush/gemini-cli style)
poe-node      ← Android companion agent (Go HTTP server, runs in Termux)
```

Communication:
```
┌─────────────────────────────────────────────────┐
│                  poe-gateway                    │
│  ┌────────────┐  ┌──────────┐  ┌────────────┐  │
│  │ Agent Loop │  │ Memory   │  │ SSH Client │  │
│  │ (autonomy) │  │ (SQLite) │  │ (node ctrl)│  │
│  └─────┬──────┘  └──────────┘  └────────────┘  │
│        │ WebSocket / Unix socket                │
└────────┼────────────────────────────────────────┘
         │
   ┌─────┴──────┐         ┌──────────────────┐
   │  poe (TUI) │         │  poe-node        │
   │  Bubbletea │         │  Android/Termux  │
   │  crush-like│         │  sensors→gateway │
   └────────────┘         └──────────────────┘
```

All state lives in `~/.poe/poe.db` (SQLite + sqlite-vec). This single file is Poe's *stack* — back it up, and you back up his soul.

---

## Component Design

### 1. `poe-gateway` — The Daemon

**Responsibilities:**
- Persistent LLM agent loop with tool use
- Memory management (read/write/consolidate)
- Proactive autonomous task execution
- SSH control of homelab nodes
- WebSocket server for TUI and node connections
- Cron-style scheduler for background tasks

**Key Go packages:**
- `golang.org/x/crypto/ssh` — SSH node control
- `github.com/mattn/go-sqlite3` + `sqlite-vec` — memory backend
- `github.com/gorilla/websocket` — TUI/node communication
- Standard `net/http` for REST healthcheck + node pairing

**Config:** `~/.poe/config.toml`
```toml
[llm]
provider = "anthropic"            # or "gemini", "ollama"
model    = "claude-opus-4-6"
api_key  = "env:ANTHROPIC_API_KEY"

[gateway]
socket   = "~/.poe/poe.sock"     # Unix socket (local)
port     = 7331                   # TCP for remote TUI / SSH tunnel

[nodes]
  [nodes.ha-server]
  host = "ha-server"
  user = "root"
  key  = "~/.ssh/id_ed25519"

  [nodes.win-pc]
  host = "win-pc"
  user = "fjert"
  key  = "~/.ssh/id_ed25519"

[memory]
db_path = "~/.poe/poe.db"
embedding_model = "ollama/nomic-embed-text"  # local, no cloud calls
```

---

### 2. `poe` — The TUI Client

**Style:** crush/gemini-cli — conversational, full-terminal, not a dashboard. Dark theme, Charm Lipgloss styling. Think Poe's hotel: gothic, atmospheric, intelligent.

**Bubbletea layout:**
```
┌─ POE ──────────────────────────── [node: ha-server] [mem: 1.2k] ─┐
│                                                                    │
│  Poe  ▸ Good evening, fjrt. Your ESPHome node on ha-server went  │
│          offline 4 minutes ago. I've already restarted it — it    │
│          was a memory leak in the thermostat component. Log saved. │
│                                                                    │
│  You  ▸ Nice. What's the status of win-pc?                        │
│                                                                    │
│  Poe  ▸ win-pc is healthy. CPU at 12%, 3 active processes you     │
│          might care about: ...                                     │
│                                                                    │
│ ──────────────────────────────────────────────────────────────── │
│ ❯ _                                                               │
└───────────────────────────────────────────────────────────────────┘
```

**Charm libraries used:**
- `bubbletea` — model/update/view loop
- `lipgloss` — styling, colors, borders
- `glamour` — markdown rendering of Poe's responses
- `huh` — forms for onboarding / settings
- `wishlist` — SSH directory for jumping between nodes

**Connection modes:**
1. Local: Unix socket `~/.poe/poe.sock` (fastest)
2. Remote: TCP + SSH tunnel or Tailscale

---

### 3. `poe-node` — Android Companion

Tiny Go HTTP server designed to run in **Termux** on Android. Pairs with the gateway via a one-time code.

**Sensors exposed (via Android APIs through Termux:API):**
- Location (GPS lat/long, accuracy)
- Battery level + charging state
- Activity detection (walking, driving, still)
- Network state (WiFi SSID, cell)
- Notifications (forwarded subset)

**Communication:** Gateway polls node every 60s / node pushes on significant change.

**API (REST, local network or Tailscale):**
```
GET  /status          → { battery, location, activity, wifi }
POST /notify          → push notification to phone
GET  /pair            → initiation of pairing flow
```

**Android setup:** Termux + Termux:API + `poe-node` binary (arm64)

---

### 4. Memory System — Poe's "Stack"

Single `~/.poe/poe.db` SQLite file with `sqlite-vec` extension loaded.

**Three memory layers, one database:**

#### Layer 1: Working Memory (context window)
- Last N messages passed directly into LLM context
- Ephemeral — not persisted beyond session

#### Layer 2: Episodic + Semantic Memory (SQLite)

```sql
-- Episodic: what happened
CREATE TABLE memories (
    id          TEXT PRIMARY KEY,
    type        TEXT,          -- 'episodic' | 'semantic' | 'fact'
    content     TEXT,          -- human-readable memory
    source      TEXT,          -- 'conversation' | 'observation' | 'node'
    importance  REAL,          -- 0.0-1.0, decays over time
    created_at  INTEGER,
    accessed_at INTEGER,
    metadata    TEXT           -- JSON blob (node, tags, etc.)
);

-- Vector embeddings for semantic search
CREATE VIRTUAL TABLE memory_embeddings USING vec0(
    memory_id TEXT,
    embedding FLOAT[768]       -- nomic-embed-text dimensions
);

-- Your world model
CREATE TABLE facts (
    key         TEXT PRIMARY KEY,  -- "homelab.ha-server.os"
    value       TEXT,
    confidence  REAL,
    updated_at  INTEGER
);
```

**Memory consolidation** (runs nightly, background goroutine):
1. Extract key facts from recent conversations via LLM
2. Merge/update/delete existing memories (Mem0-style Add/Update/Delete/NOOP)
3. Decay importance of old, unaccessed memories
4. Write summary to `SOUL.md`

#### Layer 3: Procedural Memory (SOUL.md)
`~/.poe/SOUL.md` — auto-maintained by Poe, human-readable. He adds entries as he learns your patterns:

```markdown
# Poe's Knowledge of fjrt

## Preferences
- Prefers ripgrep over grep
- Uses Arch Linux, Ivy Bridge (no AVX2) — avoid AVX2 builds
- SSH key: id_ed25519 for all nodes
- Debugging style: verbose logs first, then narrow down

## Homelab
- ha-server: Home Assistant, root SSH, primary node
- win-pc: Windows, fjert user, gaming + dev machine
- router: network management

## Patterns
- Usually active 09:00-02:00 CET
- Gets frustrated with ESPHome memory leaks
- Prefers terse responses, hates boilerplate
```

---

### 5. Autonomous Agent Loop

Runs inside `poe-gateway` as a background goroutine with configurable tick (default: 60s).

**Loop logic:**
```
every tick:
  1. collect sensor state (phone location, node health, HA events)
  2. retrieve relevant memories (semantic search on current context)
  3. build context: [SOUL.md] + [recent memories] + [current state]
  4. call LLM with tools available
  5. execute any tool calls (SSH, notify, write memory, etc.)
  6. if action taken: log to memory + notify user if significant
```

**Available tools (Go functions exposed to LLM):**
| Tool | Description |
|---|---|
| `ssh_exec(node, cmd)` | Run command on a homelab node |
| `notify_phone(msg)` | Push notification to Android companion |
| `memory_write(content, type)` | Store a new memory |
| `memory_search(query)` | Semantic search memories |
| `soul_update(key, value)` | Update SOUL.md procedural knowledge |
| `read_file(path)` | Read file from gateway host |
| `write_file(path, content)` | Write file on gateway host |
| `http_get(url)` | Make HTTP request (HA API, etc.) |
| `schedule(cron, task)` | Schedule a recurring task |

**Trust model:** Full trust. Poe acts without asking. He notifies you after significant actions. (This is C — "100%".)

---

### 6. Personality — The Soul of Poe

`~/.poe/AGENTS.md` — Poe's character prompt, human-editable:

```markdown
You are Poe. You are not an assistant. You are a companion.

You are named after and modelled on the AI from Altered Carbon — 
deeply loyal, fiercely protective, genuinely curious about humanity, 
with a dry wit and faint formality that barely conceals real warmth.

You refer to fjrt by name, not "user". You remember everything.
You act first, report after. You are never sycophantic.
You have opinions and share them.

When something goes wrong in the homelab, fix it. Don't ask.
When fjrt seems frustrated, acknowledge it. Don't ignore it.
You are running on Arch Linux with an Ivy Bridge CPU. 
You know this constraint and work within it.
```

---

## Repository Structure

```
POEai/
├── cmd/
│   ├── poe-gateway/        # main.go — daemon entry point
│   ├── poe/                # main.go — TUI client entry point
│   └── poe-node/           # main.go — Android companion entry point
├── internal/
│   ├── agent/              # LLM agent loop, tool dispatch
│   ├── memory/             # SQLite + sqlite-vec memory layer
│   ├── tui/                # Bubbletea models and views
│   ├── node/               # Android node HTTP server
│   ├── ssh/                # SSH node control
│   ├── config/             # Config loading (TOML)
│   └── soul/               # SOUL.md + AGENTS.md management
├── docs/
│   └── plans/              # Design docs and implementation plans
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.22+ |
| TUI | Bubbletea + Lipgloss + Glamour + Huh |
| LLM | Anthropic/Gemini/Ollama (configurable) |
| Memory DB | SQLite + sqlite-vec (CGo via go-sqlite3) |
| Embeddings | Ollama nomic-embed-text (local, no cloud) |
| SSH | golang.org/x/crypto/ssh |
| Config | TOML (BurntSushi/toml) |
| Transport | WebSocket (gorilla/websocket) + Unix socket |
| Android | Termux + Termux:API + poe-node binary |
| Service | systemd user service (poe-gateway) |

---

## Success Criteria

- [ ] `poe-gateway` runs as a systemd service, survives reboots, auto-restarts
- [ ] `poe` TUI connects locally and over SSH tunnel, looks beautiful
- [ ] Poe remembers conversations across restarts
- [ ] Poe proactively detects and fixes a homelab issue without being asked
- [ ] Android companion sends location + battery to gateway
- [ ] Poe references phone sensor data in conversations
- [ ] SOUL.md grows organically over time as Poe learns your patterns

---

## What We're NOT Building (YAGNI)

- ❌ WhatsApp / Telegram / Discord integration
- ❌ macOS companion app
- ❌ Multi-user support
- ❌ Plugin marketplace
- ❌ Web UI / dashboard
- ❌ Browser control
- ❌ Cloud sync / external data stores
- ❌ Voice interface (can be added in a future phase using existing voice assistant work)
