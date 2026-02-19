package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the top-level Poe configuration.
type Config struct {
	LLM     LLMConfig             `toml:"llm"`
	Gateway GatewayConfig         `toml:"gateway"`
	Memory  MemoryConfig          `toml:"memory"`
	Nodes   map[string]NodeConfig `toml:"nodes"`
}

// LLMConfig configures the language model backend.
type LLMConfig struct {
	Provider string `toml:"provider"`
	Model    string `toml:"model"`
	APIKey   string `toml:"api_key"`
}

// GatewayConfig configures the gateway daemon.
type GatewayConfig struct {
	Socket string `toml:"socket"`
	Port   int    `toml:"port"`
}

// MemoryConfig configures the memory backend.
type MemoryConfig struct {
	DBPath         string `toml:"db_path"`
	EmbeddingModel string `toml:"embedding_model"`
}

// NodeConfig configures an SSH-accessible homelab node.
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
		Nodes: make(map[string]NodeConfig),
	}
}

// Load loads configuration from the given TOML file path.
// If path is empty, returns defaults.
func Load(path string) (Config, error) {
	cfg := defaults()
	if path == "" {
		return cfg, nil
	}
	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
}
