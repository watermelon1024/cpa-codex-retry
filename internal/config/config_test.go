package config

import "testing"

func TestDecodeDefaults(t *testing.T) {
	cfg, err := Decode(nil)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if !cfg.Enabled || !cfg.InterceptStreaming || cfg.GuardRetryAttempts != 5 {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
	if cfg.StreamAction != StreamActionContinuation {
		t.Fatalf("stream action = %q", cfg.StreamAction)
	}
}

func TestDecodeRejectsDisconnectStreamAction(t *testing.T) {
	_, err := Decode([]byte("stream_action: disconnect\n"))
	if err == nil {
		t.Fatal("Decode() error = nil, want unsupported stream_action")
	}
}

func TestSupportsModelWithPrefixes(t *testing.T) {
	cfg := Default()
	cfg.Models = []string{"exact"}
	cfg.ModelPrefixes = []string{"gpt-5."}
	if !SupportsModel(cfg, "gpt-5.5") || !SupportsModel(cfg, "exact") {
		t.Fatal("SupportsModel() did not match expected model")
	}
	if SupportsModel(cfg, "claude-sonnet") {
		t.Fatal("SupportsModel() matched unexpected model")
	}
}
