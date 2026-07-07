package runner

import (
	"context"
	"fmt"
	"net/http"

	"github.com/watermelon1024/cpa-codex-retry/internal/cliproxy"
	"github.com/watermelon1024/cpa-codex-retry/internal/config"
	"github.com/watermelon1024/cpa-codex-retry/internal/guard"
	"github.com/watermelon1024/cpa-codex-retry/internal/metrics"
)

type NonStreamRunner struct {
	Config  config.Config
	Host    Host
	Metrics MetricsRecorder
}

func (r NonStreamRunner) Run(ctx context.Context, exec cliproxy.ExecutorRequest, callbackID string) (cliproxy.ExecutorResponse, *PluginError) {
	record := metrics.RequestRecord{Model: exec.Model, Format: exec.Format, Stream: false}
	defer func() {
		if r.Metrics != nil {
			r.Metrics.Record(record)
		}
	}()

	body := requestBody(exec)
	effort := requestEffort(body)
	for attempt := 0; attempt <= r.Config.GuardRetryAttempts; attempt++ {
		record.Attempts = attempt + 1
		resp, pluginErr := r.executeAttempt(ctx, exec, callbackID, body)
		if pluginErr != nil {
			record.ErrorCode = pluginErr.Code
			record.HTTPStatus = pluginErr.HTTPStatus
			if r.shouldRetryCapacity(pluginErr, nil, attempt) {
				record.RetryAttempts++
				continue
			}
			return cliproxy.ExecutorResponse{}, pluginErr
		}
		decision := r.inspect(resp, exec, body, effort)
		markDecisionObservation(&record, decision)
		if !decision.Matched || !r.Config.InterceptNonStreaming {
			return executorResponse(resp), nil
		}
		markGuardMatch(&record, decision)
		if attempt < r.Config.GuardRetryAttempts {
			record.RetryAttempts++
			r.log(ctx, callbackID, "info", "non-stream reasoning guard retry", logFields(attempt, decision))
			continue
		}
		record.Blocked = true
		record.HTTPStatus = r.Config.NonStreamStatusCode
		return cliproxy.ExecutorResponse{}, blockedError(r.Config.NonStreamStatusCode, blockMessage(decision))
	}
	record.Blocked = true
	record.HTTPStatus = http.StatusBadGateway
	return cliproxy.ExecutorResponse{}, blockedError(http.StatusBadGateway, "retry loop exhausted")
}

func (r NonStreamRunner) executeAttempt(ctx context.Context, exec cliproxy.ExecutorRequest, callbackID string, body []byte) (cliproxy.HostModelExecutionResponse, *PluginError) {
	resp, err := r.Host.ExecuteModel(ctx, hostRequest(exec, callbackID, body, false))
	if err != nil {
		return cliproxy.HostModelExecutionResponse{}, upstreamError(err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		if r.shouldRetryCapacity(nil, resp.Body, 0) {
			return resp, &PluginError{Code: "capacity", Message: capacityErrorText, HTTPStatus: resp.StatusCode}
		}
		return resp, statusError(resp.StatusCode, string(resp.Body))
	}
	return resp, nil
}

func (r NonStreamRunner) inspect(resp cliproxy.HostModelExecutionResponse, exec cliproxy.ExecutorRequest, body []byte, effort string) guard.Decision {
	inspection := guard.InspectJSON(resp.Body, exec.Headers, body)
	return guard.Match(r.Config, inspection, effort)
}

func (r NonStreamRunner) shouldRetryCapacity(pluginErr *PluginError, body []byte, attempt int) bool {
	if !r.Config.RetryUpstreamCapacityErrors || attempt >= r.Config.GuardRetryAttempts {
		return false
	}
	if pluginErr != nil {
		return isCapacityError(fmt.Errorf("%s", pluginErr.Message), nil)
	}
	return isCapacityError(nil, body)
}

func (r NonStreamRunner) log(ctx context.Context, callbackID, level, message string, fields map[string]any) {
	if r.Host == nil || !r.Config.LogMatch {
		return
	}
	if fields == nil {
		fields = map[string]any{}
	}
	fields["host_callback_id"] = callbackID
	r.Host.Log(ctx, level, message, fields)
}

func executorResponse(resp cliproxy.HostModelExecutionResponse) cliproxy.ExecutorResponse {
	return cliproxy.ExecutorResponse{
		Payload: append([]byte(nil), resp.Body...),
		Headers: cloneHeader(resp.Headers),
	}
}

func blockMessage(decision guard.Decision) string {
	if decision.BlockedReasoning == nil {
		return "codex retry gateway blocked suspicious reasoning response"
	}
	return fmt.Sprintf("codex retry gateway blocked reasoning_tokens=%d", *decision.BlockedReasoning)
}

func logFields(attempt int, decision guard.Decision) map[string]any {
	return map[string]any{
		"attempt": attempt,
		"mode":    decision.Mode,
		"reason":  decision.Reason,
	}
}

func markGuardMatch(record *metrics.RequestRecord, decision guard.Decision) {
	record.Intercepted = true
	record.GuardMatches++
	record.Mode = decision.Mode
	record.Reason = decision.Reason
	if decision.ReasoningTokens != nil {
		value := *decision.ReasoningTokens
		record.ReasoningToken = &value
	}
	if decision.BlockedReasoning != nil {
		value := *decision.BlockedReasoning
		record.ReasoningToken = &value
	}
}

func markDecisionObservation(record *metrics.RequestRecord, decision guard.Decision) {
	if record.Mode == "" {
		record.Mode = decision.Mode
	}
	if record.ReasoningToken != nil || decision.ReasoningTokens == nil {
		return
	}
	value := *decision.ReasoningTokens
	record.ReasoningToken = &value
}
