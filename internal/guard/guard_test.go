package guard

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/watermelon1024/cpa-codex-retry/internal/config"
)

func TestMatchFormulaReasoningTokens(t *testing.T) {
	cfg := config.Default()
	reasoning := 1034
	decision := Match(cfg, Inspection{ReasoningTokens: &reasoning}, "")
	if !decision.Matched || decision.BlockedReasoning == nil {
		t.Fatalf("Match() = %#v", decision)
	}
}

func TestInspectJSONFindsInteractionsReasoningTokens(t *testing.T) {
	inspection := InspectJSON([]byte(`{"event_type":"finish","metadata":{"total_usage":{"total_thought_tokens":516}}}`), nil, nil)
	if inspection.ReasoningTokens == nil || *inspection.ReasoningTokens != 516 {
		t.Fatalf("ReasoningTokens = %#v, want 516", inspection.ReasoningTokens)
	}
}

func TestInspectJSONFindsTopLevelReasoningTokens(t *testing.T) {
	inspection := InspectJSON([]byte(`{"usage":{"reasoning_tokens":516}}`), nil, nil)
	if inspection.ReasoningTokens == nil || *inspection.ReasoningTokens != 516 {
		t.Fatalf("ReasoningTokens = %#v, want 516", inspection.ReasoningTokens)
	}
}

func TestContextCompactionZeroExempt(t *testing.T) {
	cfg := config.Default()
	reasoning := 0
	decision := Match(cfg, Inspection{
		ReasoningTokens: &reasoning,
		RequestKind:     RequestKindContextCompaction,
	}, "")
	if !decision.Exempt || decision.Matched {
		t.Fatalf("Match() = %#v", decision)
	}
}

func TestFinalAnswerOnlyHighEffort(t *testing.T) {
	cfg := config.Default()
	cfg.InterceptRuleMode = RuleFinalOnlyHighXHigh
	reasoning := 7
	decision := Match(cfg, Inspection{
		ReasoningTokens: &reasoning,
		Structure:       Structure{HasFinalAnswer: true, HasOutputText: true},
	}, "high")
	if !decision.Matched {
		t.Fatalf("Match() = %#v", decision)
	}
}

func TestBuildContinuationRequestSanitizesReplay(t *testing.T) {
	cfg := config.Default()
	base := []byte(`{"previous_response_id":"r","include":["reasoning.encrypted_content"],"input":[{"type":"reasoning","encrypted_content":"x"},{"role":"user","content":"hi","encrypted_content":"secret"}]}`)
	next, err := BuildContinuationRequest(cfg, base)
	if err != nil {
		t.Fatalf("BuildContinuationRequest() error = %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(next, &payload); err != nil {
		t.Fatalf("decode continuation: %v", err)
	}
	if _, ok := payload["previous_response_id"]; ok {
		t.Fatal("previous_response_id was not removed")
	}
	if _, ok := payload["include"]; ok {
		t.Fatal("encrypted include was not removed")
	}
	input := payload["input"].([]any)
	if len(input) != 2 {
		t.Fatalf("input len = %d, want 2", len(input))
	}
	if containsKey(input[0].(map[string]any), "encrypted_content") {
		t.Fatal("encrypted_content was not stripped")
	}
}

func TestInspectJSONDetectsRequestKind(t *testing.T) {
	headers := http.Header{"X-Codex-Purpose": []string{"context_compaction"}}
	inspection := InspectJSON([]byte(`{}`), headers, []byte(`{}`))
	if inspection.RequestKind != RequestKindContextCompaction {
		t.Fatalf("request kind = %q", inspection.RequestKind)
	}
}

func containsKey(values map[string]any, key string) bool {
	_, ok := values[key]
	return ok
}
