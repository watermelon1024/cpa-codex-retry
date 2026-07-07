package runner

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/watermelon1024/cpa-codex-retry/internal/cliproxy"
	"github.com/watermelon1024/cpa-codex-retry/internal/config"
	"github.com/watermelon1024/cpa-codex-retry/internal/metrics"
)

type fakeHost struct {
	executeResponses []cliproxy.HostModelExecutionResponse
	streamResponses  []fakeStream
	executeBodies    [][]byte
	streamBodies     [][]byte
	activeChunks     []cliproxy.HostModelStreamReadResponse
}

type fakeStream struct {
	headers http.Header
	chunks  []cliproxy.HostModelStreamReadResponse
	err     error
}

type captureMetrics struct {
	records []metrics.RequestRecord
}

func (m *captureMetrics) Record(record metrics.RequestRecord) {
	m.records = append(m.records, record)
}

func (h *fakeHost) ExecuteModel(_ context.Context, req cliproxy.RPCHostModelExecutionRequest) (cliproxy.HostModelExecutionResponse, error) {
	h.executeBodies = append(h.executeBodies, append([]byte(nil), req.Body...))
	if len(h.executeResponses) == 0 {
		return cliproxy.HostModelExecutionResponse{}, fmt.Errorf("no execute response")
	}
	resp := h.executeResponses[0]
	h.executeResponses = h.executeResponses[1:]
	return resp, nil
}

func (h *fakeHost) ExecuteModelStream(_ context.Context, req cliproxy.RPCHostModelExecutionRequest) (cliproxy.HostModelStreamResponse, error) {
	h.streamBodies = append(h.streamBodies, append([]byte(nil), req.Body...))
	if len(h.streamResponses) == 0 {
		return cliproxy.HostModelStreamResponse{}, fmt.Errorf("no stream response")
	}
	current := h.streamResponses[0]
	h.streamResponses = h.streamResponses[1:]
	h.activeChunks = append([]cliproxy.HostModelStreamReadResponse(nil), current.chunks...)
	if current.err != nil {
		return cliproxy.HostModelStreamResponse{}, current.err
	}
	return cliproxy.HostModelStreamResponse{
		StatusCode: http.StatusOK,
		Headers:    current.headers,
		StreamID:   "stream",
	}, nil
}

func (h *fakeHost) ReadModelStream(_ context.Context, _ string) (cliproxy.HostModelStreamReadResponse, error) {
	if len(h.activeChunks) == 0 {
		return cliproxy.HostModelStreamReadResponse{Done: true}, nil
	}
	chunk := h.activeChunks[0]
	h.activeChunks = h.activeChunks[1:]
	return chunk, nil
}

func (h *fakeHost) CloseModelStream(context.Context, string) error {
	return nil
}

func (h *fakeHost) Log(context.Context, string, string, map[string]any) {}

func TestNonStreamRetriesReasoningMatch(t *testing.T) {
	metricSink := &captureMetrics{}
	host := &fakeHost{executeResponses: []cliproxy.HostModelExecutionResponse{
		jsonResponse(`{"usage":{"output_tokens_details":{"reasoning_tokens":516}}}`),
		jsonResponse(`{"usage":{"output_tokens_details":{"reasoning_tokens":1}}}`),
	}}
	resp, err := (NonStreamRunner{Config: config.Default(), Host: host, Metrics: metricSink}).Run(context.Background(), execRequest(false), "cb")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if string(resp.Payload) == "" || len(host.executeBodies) != 2 {
		t.Fatalf("payload=%q attempts=%d", resp.Payload, len(host.executeBodies))
	}
	if len(metricSink.records) != 1 {
		t.Fatalf("metrics records len = %d, want 1", len(metricSink.records))
	}
	record := metricSink.records[0]
	if !record.Intercepted || record.Blocked || record.Attempts != 2 || record.RetryAttempts != 1 || record.GuardMatches != 1 {
		t.Fatalf("metrics record = %#v", record)
	}
	if record.ReasoningToken == nil || *record.ReasoningToken != 516 {
		t.Fatalf("reasoning token = %#v, want matched token 516", record.ReasoningToken)
	}
}

func TestNonStreamRetriesInteractionsReasoningMatch(t *testing.T) {
	metricSink := &captureMetrics{}
	host := &fakeHost{executeResponses: []cliproxy.HostModelExecutionResponse{
		jsonResponse(`{"event_type":"finish","metadata":{"total_usage":{"total_thought_tokens":516}}}`),
		jsonResponse(`{"event_type":"finish","metadata":{"total_usage":{"total_thought_tokens":1}}}`),
	}}
	_, err := (NonStreamRunner{Config: config.Default(), Host: host, Metrics: metricSink}).Run(context.Background(), execRequest(false), "cb")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(host.executeBodies) != 2 || len(metricSink.records) != 1 {
		t.Fatalf("attempts=%d metrics=%d, want 2 attempts and 1 metric", len(host.executeBodies), len(metricSink.records))
	}
	record := metricSink.records[0]
	if !record.Intercepted || record.RetryAttempts != 1 || record.ReasoningToken == nil || *record.ReasoningToken != 516 {
		t.Fatalf("metrics record = %#v", record)
	}
}

func TestNonStreamRecordsReasoningTokensWithoutMatch(t *testing.T) {
	metricSink := &captureMetrics{}
	host := &fakeHost{executeResponses: []cliproxy.HostModelExecutionResponse{
		jsonResponse(`{"usage":{"reasoning_tokens":12}}`),
	}}
	_, err := (NonStreamRunner{Config: config.Default(), Host: host, Metrics: metricSink}).Run(context.Background(), execRequest(false), "cb")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(metricSink.records) != 1 {
		t.Fatalf("metrics records len = %d, want 1", len(metricSink.records))
	}
	record := metricSink.records[0]
	if record.Intercepted || record.ReasoningToken == nil || *record.ReasoningToken != 12 {
		t.Fatalf("metrics record = %#v", record)
	}
}

func TestNonStreamBlocksAfterRetryLimit(t *testing.T) {
	cfg := config.Default()
	cfg.GuardRetryAttempts = 0
	host := &fakeHost{executeResponses: []cliproxy.HostModelExecutionResponse{
		jsonResponse(`{"usage":{"output_tokens_details":{"reasoning_tokens":516}}}`),
	}}
	_, err := (NonStreamRunner{Config: cfg, Host: host}).Run(context.Background(), execRequest(false), "cb")
	if err == nil || err.HTTPStatus != http.StatusBadGateway {
		t.Fatalf("Run() error = %#v", err)
	}
}

func TestStreamContinuationRecovery(t *testing.T) {
	host := &fakeHost{streamResponses: []fakeStream{
		{chunks: []cliproxy.HostModelStreamReadResponse{
			sseChunk(`{"type":"response.completed","response":{"usage":{"output_tokens_details":{"reasoning_tokens":516}}}}`, false),
		}},
		{chunks: []cliproxy.HostModelStreamReadResponse{
			sseChunk(`{"type":"response.output_text.delta","delta":"ok"}`, false),
			{Done: true},
		}},
	}}
	resp, err := (StreamRunner{Config: config.Default(), Host: host}).Run(context.Background(), execRequest(true), "cb")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(resp.Chunks) != 1 || len(host.streamBodies) != 2 {
		t.Fatalf("chunks=%d attempts=%d", len(resp.Chunks), len(host.streamBodies))
	}
	if string(host.streamBodies[1]) == string(host.streamBodies[0]) {
		t.Fatal("second stream attempt did not use continuation body")
	}
}

func jsonResponse(body string) cliproxy.HostModelExecutionResponse {
	return cliproxy.HostModelExecutionResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(body),
	}
}

func execRequest(stream bool) cliproxy.ExecutorRequest {
	return cliproxy.ExecutorRequest{
		Model:        "gpt-5.5",
		Format:       "codex",
		SourceFormat: "codex",
		Stream:       stream,
		Payload:      []byte(`{"model":"gpt-5.5","stream":true,"input":[{"role":"user","content":"hi"}],"reasoning":{"effort":"high"}}`),
	}
}

func sseChunk(payload string, done bool) cliproxy.HostModelStreamReadResponse {
	return cliproxy.HostModelStreamReadResponse{
		Payload: []byte("data: " + payload + "\n\n"),
		Done:    done,
	}
}
