package runner

import (
	"context"
	"fmt"
	"net/http"

	"github.com/watermelon1024/cpa-codex-retry/internal/cliproxy"
	"github.com/watermelon1024/cpa-codex-retry/internal/config"
	"github.com/watermelon1024/cpa-codex-retry/internal/guard"
)

type StreamRunner struct {
	Config config.Config
	Host   Host
}

type streamAttempt struct {
	Headers  http.Header
	Chunks   [][]byte
	Decision guard.Decision
}

func (r StreamRunner) Run(ctx context.Context, exec cliproxy.ExecutorRequest, callbackID string) (cliproxy.ExecutorStreamResponse, *PluginError) {
	baseBody := requestBody(exec)
	body := append([]byte(nil), baseBody...)
	effort := requestEffort(baseBody)
	for attempt := 0; attempt <= r.Config.GuardRetryAttempts; attempt++ {
		result, pluginErr := r.streamAttempt(ctx, exec, callbackID, body, effort)
		if pluginErr != nil {
			if r.shouldRetryCapacity(pluginErr, nil, attempt) {
				continue
			}
			return cliproxy.ExecutorStreamResponse{}, pluginErr
		}
		nextBody, retry := r.nextStreamBody(ctx, callbackID, baseBody, result, attempt)
		if retry {
			body = nextBody
			continue
		}
		if result.Decision.Matched && r.Config.InterceptStreaming {
			return cliproxy.ExecutorStreamResponse{}, blockedError(r.Config.NonStreamStatusCode, blockMessage(result.Decision))
		}
		return streamResponse(result.Headers, result.Chunks, r.Config.StripEncryptedContent), nil
	}
	return cliproxy.ExecutorStreamResponse{}, blockedError(http.StatusBadGateway, "stream retry loop exhausted")
}

func (r StreamRunner) streamAttempt(ctx context.Context, exec cliproxy.ExecutorRequest, callbackID string, body []byte, effort string) (streamAttempt, *PluginError) {
	resp, err := r.Host.ExecuteModelStream(ctx, hostRequest(exec, callbackID, body, true))
	if err != nil {
		return streamAttempt{}, upstreamError(err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return streamAttempt{}, statusError(resp.StatusCode, "")
	}
	if resp.StreamID == "" {
		return streamAttempt{}, upstreamError(fmt.Errorf("host model stream returned empty stream_id"))
	}
	return r.readStream(ctx, resp, exec, body, effort)
}

func (r StreamRunner) readStream(ctx context.Context, resp cliproxy.HostModelStreamResponse, exec cliproxy.ExecutorRequest, body []byte, effort string) (streamAttempt, *PluginError) {
	defer func() { _ = r.Host.CloseModelStream(ctx, resp.StreamID) }()
	state := guard.InspectJSON(nil, exec.Headers, body)
	parser := &guard.SSEParser{}
	chunks := make([][]byte, 0)
	for {
		chunk, errRead := r.Host.ReadModelStream(ctx, resp.StreamID)
		if errRead != nil {
			return streamAttempt{}, upstreamError(errRead)
		}
		if chunk.Error != "" {
			return streamAttempt{}, upstreamError(fmt.Errorf("%s", chunk.Error))
		}
		if len(chunk.Payload) > 0 {
			state, _ = parser.Push(chunk.Payload, state)
			chunks = append(chunks, append([]byte(nil), chunk.Payload...))
			if decision := guard.Match(r.Config, state, effort); earlyStreamMatch(decision) {
				return streamAttempt{Headers: resp.Headers, Chunks: chunks, Decision: decision}, nil
			}
		}
		if chunk.Done {
			state = parser.Flush(state)
			decision := guard.Match(r.Config, state, effort)
			return streamAttempt{Headers: resp.Headers, Chunks: chunks, Decision: decision}, nil
		}
	}
}

func (r StreamRunner) nextStreamBody(ctx context.Context, callbackID string, baseBody []byte, result streamAttempt, attempt int) ([]byte, bool) {
	if !result.Decision.Matched || !r.Config.InterceptStreaming {
		return nil, false
	}
	if attempt >= r.Config.GuardRetryAttempts {
		return nil, false
	}
	r.log(ctx, callbackID, "info", "stream reasoning guard retry", logFields(attempt, result.Decision))
	if !canContinuationRecover(r.Config, result.Decision, baseBody) {
		return append([]byte(nil), baseBody...), true
	}
	next, err := guard.BuildContinuationRequest(r.Config, baseBody)
	if err != nil {
		return append([]byte(nil), baseBody...), true
	}
	return next, true
}

func (r StreamRunner) shouldRetryCapacity(pluginErr *PluginError, body []byte, attempt int) bool {
	if !r.Config.RetryUpstreamCapacityErrors || attempt >= r.Config.GuardRetryAttempts {
		return false
	}
	if pluginErr != nil {
		return isCapacityError(fmt.Errorf("%s", pluginErr.Message), nil)
	}
	return isCapacityError(nil, body)
}

func (r StreamRunner) log(ctx context.Context, callbackID, level, message string, fields map[string]any) {
	if r.Host == nil || !r.Config.LogMatch {
		return
	}
	if fields == nil {
		fields = map[string]any{}
	}
	fields["host_callback_id"] = callbackID
	r.Host.Log(ctx, level, message, fields)
}

func earlyStreamMatch(decision guard.Decision) bool {
	return decision.Matched && decision.Mode == guard.RuleReasoningTokens
}

func canContinuationRecover(cfg config.Config, decision guard.Decision, baseBody []byte) bool {
	return cfg.StreamAction == config.StreamActionContinuation &&
		decision.Mode == guard.RuleReasoningTokens &&
		guard.IsResponsesRequest(baseBody)
}

func streamResponse(headers http.Header, chunks [][]byte, stripEncrypted bool) cliproxy.ExecutorStreamResponse {
	out := make([]cliproxy.ExecutorStreamChunk, 0, len(chunks))
	for _, chunk := range chunks {
		payload := append([]byte(nil), chunk...)
		if stripEncrypted {
			payload = guard.StripEncryptedContentSSE(payload)
		}
		out = append(out, cliproxy.ExecutorStreamChunk{Payload: payload})
	}
	return cliproxy.ExecutorStreamResponse{Headers: cloneHeader(headers), Chunks: out}
}
