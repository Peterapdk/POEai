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
	if cfg.Memory.DBPath == "" {
		t.Error("DBPath should not be empty")
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
