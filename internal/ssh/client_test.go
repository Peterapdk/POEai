package ssh_test

import (
	"context"
	"testing"

	"github.com/fjrt/poeai/internal/ssh"
)

func TestClient_Exec_Error(t *testing.T) {
	// Simple test that we can instantiate and it fails correctly with non-existent key
	client := ssh.New("localhost", "fjrt", "/nonexistent/key")
	_, err := client.Exec(context.Background(), "echo hello")
	if err == nil {
		t.Error("Exec() with non-existent key should fail")
	}
}
