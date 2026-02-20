# Poe AI Sidekick

Autonomous, evolving, proactive AI sidekick inspired by Altered Carbon.

## Components
- **poe-gateway**: Persistent daemon on your homelab.
- **poe**: Charm-powered TUI client for terminal-first interaction.
- **poe-node**: Android companion agent (running in Termux) for sensor data.

## Getting Started

1. **Build**:
   ```bash
   make build
   ```

2. **Run Gateway**:
   ```bash
   ./poe-gateway
   ```

3. **Connect TUI**:
   ```bash
   ./poe
   ```

4. **Android Setup**:
   Install Termux, compile `poe-node` for arm64, and run it.

## Architecture
Built with Go, Bubbletea, and SQLite. Uses human-like memory layers (episodic, semantic, procedural).
