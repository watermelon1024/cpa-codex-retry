package runner

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/watermelon1024/cpa-codex-retry/internal/cliproxy"
)

func hostRequest(exec cliproxy.ExecutorRequest, callbackID string, body []byte, stream bool) cliproxy.RPCHostModelExecutionRequest {
	entry := strings.TrimSpace(exec.SourceFormat)
	exit := strings.TrimSpace(exec.Format)
	if entry == "" {
		entry = exit
	}
	if exit == "" {
		exit = entry
	}
	return cliproxy.RPCHostModelExecutionRequest{
		HostModelExecutionRequest: cliproxy.HostModelExecutionRequest{
			EntryProtocol: entry,
			ExitProtocol:  exit,
			Model:         exec.Model,
			Stream:        stream,
			Body:          bodyOrFallback(body, exec),
			Headers:       cloneHeader(exec.Headers),
			Query:         exec.Query,
			Alt:           exec.Alt,
		},
		HostCallbackID: callbackID,
	}
}

func requestBody(exec cliproxy.ExecutorRequest) []byte {
	if len(exec.Payload) > 0 {
		return append([]byte(nil), exec.Payload...)
	}
	return append([]byte(nil), exec.OriginalRequest...)
}

func requestEffort(body []byte) string {
	var payload map[string]any
	if json.Unmarshal(body, &payload) != nil {
		return ""
	}
	reasoning, ok := payload["reasoning"].(map[string]any)
	if !ok {
		return ""
	}
	effort, _ := reasoning["effort"].(string)
	return strings.ToLower(strings.TrimSpace(effort))
}

func bodyOrFallback(body []byte, exec cliproxy.ExecutorRequest) []byte {
	if len(body) > 0 {
		return append([]byte(nil), body...)
	}
	return requestBody(exec)
}

func cloneHeader(headers http.Header) http.Header {
	if headers == nil {
		return nil
	}
	return headers.Clone()
}
