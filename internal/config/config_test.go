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
	if !SupportsSourceFormat(cfg, "interactions") {
		t.Fatalf("default source formats = %#v, want interactions support", cfg.SourceFormats)
	}
}

func TestDecodeRejectsDisconnectStreamAction(t *testing.T) {
	_, err := Decode([]byte("stream_action: disconnect\n"))
	if err == nil {
		t.Fatal("Decode() error = nil, want unsupported stream_action")
	}
}

func TestDecodeReasoningEqualsAcceptsStringInts(t *testing.T) {
	cfg, err := Decode([]byte(`
reasoning_equals:
  - "516"
  - "1034"
  - 1552
  - "2070"
`))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	want := []int{516, 1034, 1552, 2070}
	if !equalInts(cfg.ReasoningEquals, want) {
		t.Fatalf("ReasoningEquals = %#v, want %#v", cfg.ReasoningEquals, want)
	}
}

func TestDecodeReasoningEqualsRejectsInvalidString(t *testing.T) {
	_, err := Decode([]byte(`
reasoning_equals:
  - "516"
  - nope
`))
	if err == nil {
		t.Fatal("Decode() error = nil, want invalid reasoning_equals")
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

func equalInts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
