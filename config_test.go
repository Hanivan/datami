package datami

import (
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Config{}
	if cfg.Timeout != 0 {
		t.Fatal("expected zero Timeout on empty config")
	}
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New with empty config should not fail: %v", err)
	}
	if client == nil {
		t.Fatal("client should not be nil")
	}
}

func TestConfigTimeoutDefault(t *testing.T) {
	cfg := Config{Timeout: 5 * time.Second}
	if cfg.Timeout != 5*time.Second {
		t.Fatal("Timeout not set")
	}
}
