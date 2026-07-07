package runner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/watermelon1024/cpa-codex-retry/internal/cliproxy"
)

const capacityErrorText = "Selected model is at capacity. Please try a different model."

type Host interface {
	ExecuteModel(context.Context, cliproxy.RPCHostModelExecutionRequest) (cliproxy.HostModelExecutionResponse, error)
	ExecuteModelStream(context.Context, cliproxy.RPCHostModelExecutionRequest) (cliproxy.HostModelStreamResponse, error)
	ReadModelStream(context.Context, string) (cliproxy.HostModelStreamReadResponse, error)
	CloseModelStream(context.Context, string) error
	Log(context.Context, string, string, map[string]any)
}

type PluginError struct {
	Code       string
	Message    string
	HTTPStatus int
	Retryable  bool
}

func (e *PluginError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func blockedError(status int, message string) *PluginError {
	return &PluginError{Code: "reasoning_guard_triggered", Message: message, HTTPStatus: status}
}

func upstreamError(err error) *PluginError {
	status := http.StatusBadGateway
	var statusErr interface{ StatusCode() int }
	if errors.As(err, &statusErr) && statusErr.StatusCode() > 0 {
		status = statusErr.StatusCode()
	}
	return &PluginError{Code: "upstream_error", Message: err.Error(), HTTPStatus: status}
}

func statusError(status int, message string) *PluginError {
	if strings.TrimSpace(message) == "" {
		message = fmt.Sprintf("host model status %d", status)
	}
	return &PluginError{Code: "upstream_status", Message: message, HTTPStatus: status}
}

func isCapacityError(err error, body []byte) bool {
	if err != nil && strings.Contains(err.Error(), capacityErrorText) {
		return true
	}
	return strings.Contains(string(body), capacityErrorText)
}
