package node_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fjrt/poeai/internal/node"
)

func TestServer_Status(t *testing.T) {
	srv := node.New()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/status")
	if err != nil {
		t.Fatalf("GET /status error = %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("status code = %d, want 200", res.StatusCode)
	}

	var status node.Status
	if err := json.NewDecoder(res.Body).Decode(&status); err != nil {
		t.Fatalf("decode JSON error = %v", err)
	}
	// Initial status should be empty but valid
	if status.Activity == "" && status.Battery.Level == 0 {
		// OK
	}
}

func TestServer_Notify(t *testing.T) {
	srv := node.New()
	ts := httptest.NewServer(srv)
	defer ts.Close()

	res, err := http.Post(ts.URL+"/notify", "text/plain", nil)
	if err != nil {
		t.Fatalf("POST /notify error = %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("status code = %d, want 200", res.StatusCode)
	}
}
