package main

/*
#include <stdint.h>
#include <stdlib.h>

typedef struct {
	void* ptr;
	size_t len;
} cliproxy_buffer;

typedef int (*cliproxy_host_call_fn)(void*, const char*, const uint8_t*, size_t, cliproxy_buffer*);
typedef void (*cliproxy_host_free_fn)(void*, size_t);

typedef struct {
	uint32_t abi_version;
	void* host_ctx;
	cliproxy_host_call_fn call;
	cliproxy_host_free_fn free_buffer;
} cliproxy_host_api;

typedef int (*cliproxy_plugin_call_fn)(char*, uint8_t*, size_t, cliproxy_buffer*);
typedef void (*cliproxy_plugin_free_fn)(void*, size_t);
typedef void (*cliproxy_plugin_shutdown_fn)(void);

typedef struct {
	uint32_t abi_version;
	cliproxy_plugin_call_fn call;
	cliproxy_plugin_free_fn free_buffer;
	cliproxy_plugin_shutdown_fn shutdown;
} cliproxy_plugin_api;

#ifdef _WIN32
#define CLIPROXY_EXPORT __declspec(dllexport)
#else
#define CLIPROXY_EXPORT
#endif

extern CLIPROXY_EXPORT int cliproxyPluginCall(char*, uint8_t*, size_t, cliproxy_buffer*);
extern CLIPROXY_EXPORT void cliproxyPluginFree(void*, size_t);
extern CLIPROXY_EXPORT void cliproxyPluginShutdown(void);

static const cliproxy_host_api* stored_host;

static void store_host_api(const cliproxy_host_api* host) {
	stored_host = host;
}

static int call_host_api(const char* method, const uint8_t* request, size_t request_len, cliproxy_buffer* response) {
	if (stored_host == NULL || stored_host->call == NULL) {
		return 1;
	}
	return stored_host->call(stored_host->host_ctx, method, request, request_len, response);
}

static void free_host_buffer(void* ptr, size_t len) {
	if (stored_host != NULL && stored_host->free_buffer != NULL && ptr != NULL) {
		stored_host->free_buffer(ptr, len);
	}
}
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"unsafe"

	"github.com/watermelon1024/cpa-codex-retry/internal/cliproxy"
	"github.com/watermelon1024/cpa-codex-retry/internal/config"
	"github.com/watermelon1024/cpa-codex-retry/internal/runner"
)

var activeConfig atomic.Value

func init() {
	activeConfig.Store(config.Default())
}

func main() {}

//export cliproxy_plugin_init
func cliproxy_plugin_init(host *C.cliproxy_host_api, plugin *C.cliproxy_plugin_api) C.int {
	if plugin == nil {
		return 1
	}
	C.store_host_api(host)
	plugin.abi_version = C.uint32_t(cliproxy.ABIVersion)
	plugin.call = C.cliproxy_plugin_call_fn(C.cliproxyPluginCall)
	plugin.free_buffer = C.cliproxy_plugin_free_fn(C.cliproxyPluginFree)
	plugin.shutdown = C.cliproxy_plugin_shutdown_fn(C.cliproxyPluginShutdown)
	return 0
}

//export cliproxyPluginCall
func cliproxyPluginCall(method *C.char, request *C.uint8_t, requestLen C.size_t, response *C.cliproxy_buffer) C.int {
	if response != nil {
		response.ptr = nil
		response.len = 0
	}
	if method == nil {
		writeResponse(response, errorEnvelope("invalid_method", "method is required", http.StatusBadRequest))
		return 1
	}
	raw := dispatch(C.GoString(method), requestBytes(request, requestLen))
	writeResponse(response, raw)
	return 0
}

//export cliproxyPluginFree
func cliproxyPluginFree(ptr unsafe.Pointer, _ C.size_t) {
	if ptr != nil {
		C.free(ptr)
	}
}

//export cliproxyPluginShutdown
func cliproxyPluginShutdown() {}

func dispatch(method string, request []byte) []byte {
	raw, err := handleMethod(method, request)
	if err != nil {
		return errorEnvelope("plugin_error", err.Error(), http.StatusInternalServerError)
	}
	return raw
}

func handleMethod(method string, request []byte) ([]byte, error) {
	switch method {
	case cliproxy.MethodPluginRegister, cliproxy.MethodPluginReconfigure:
		return configure(request)
	case cliproxy.MethodModelRoute:
		return routeModel(request)
	case cliproxy.MethodExecutorIdentifier:
		return okEnvelope(map[string]string{"identifier": cliproxy.PluginID})
	case cliproxy.MethodExecutorExecute:
		return execute(request)
	case cliproxy.MethodExecutorExecuteStream:
		return executeStream(request)
	case cliproxy.MethodExecutorCountTokens, cliproxy.MethodExecutorHTTPRequest:
		return errorEnvelope("not_supported", method+" is not supported by this plugin", http.StatusNotImplemented), nil
	default:
		return errorEnvelope("unknown_method", "unknown method: "+method, http.StatusNotFound), nil
	}
}

func configure(raw []byte) ([]byte, error) {
	var req cliproxy.LifecycleRequest
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, err
		}
	}
	cfg, err := config.Decode(req.ConfigYAML)
	if err != nil {
		return nil, err
	}
	activeConfig.Store(cfg)
	return okEnvelope(registration(cfg))
}

func routeModel(raw []byte) ([]byte, error) {
	var req cliproxy.ModelRouteRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}
	cfg := currentConfig()
	if !cfg.Enabled || !config.SupportsSourceFormat(cfg, req.SourceFormat) {
		return okEnvelope(cliproxy.ModelRouteResponse{Handled: false})
	}
	if !config.SupportsModel(cfg, req.RequestedModel) {
		return okEnvelope(cliproxy.ModelRouteResponse{Handled: false})
	}
	return okEnvelope(cliproxy.ModelRouteResponse{
		Handled:    true,
		TargetKind: "self",
		Reason:     "codex_retry_gateway_guard",
	})
}

func execute(raw []byte) ([]byte, error) {
	var req cliproxy.RPCExecutorRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}
	resp, pluginErr := (runner.NonStreamRunner{Config: currentConfig(), Host: cHostClient{}}).
		Run(context.Background(), req.ExecutorRequest, req.HostCallbackID)
	return executorEnvelope(resp, pluginErr)
}

func executeStream(raw []byte) ([]byte, error) {
	var req cliproxy.RPCExecutorRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}
	resp, pluginErr := (runner.StreamRunner{Config: currentConfig(), Host: cHostClient{}}).
		Run(context.Background(), req.ExecutorRequest, req.HostCallbackID)
	return executorEnvelope(resp, pluginErr)
}

func executorEnvelope(result any, pluginErr *runner.PluginError) ([]byte, error) {
	if pluginErr != nil {
		return errorEnvelope(pluginErr.Code, pluginErr.Message, pluginErr.HTTPStatus), nil
	}
	return okEnvelope(result)
}

func currentConfig() config.Config {
	cfg, ok := activeConfig.Load().(config.Config)
	if ok {
		return cfg
	}
	return config.Default()
}

func requestBytes(request *C.uint8_t, requestLen C.size_t) []byte {
	if request == nil || requestLen == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(request), C.int(requestLen))
}

func okEnvelope(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return json.Marshal(cliproxy.Envelope{OK: true, Result: raw})
}

func errorEnvelope(code, message string, status int) []byte {
	raw, _ := json.Marshal(cliproxy.Envelope{
		OK: false,
		Error: &cliproxy.Error{
			Code:       code,
			Message:    message,
			HTTPStatus: status,
		},
	})
	return raw
}

func writeResponse(response *C.cliproxy_buffer, raw []byte) {
	if response == nil || len(raw) == 0 {
		return
	}
	ptr := C.CBytes(raw)
	if ptr == nil {
		return
	}
	response.ptr = ptr
	response.len = C.size_t(len(raw))
}

func registration(_ config.Config) cliproxy.Registration {
	formats := []string{"codex", "openai-response", "openai"}
	return cliproxy.Registration{
		SchemaVersion: cliproxy.SchemaVersion,
		Metadata: cliproxy.Metadata{
			Name:             cliproxy.PluginID,
			Version:          "0.1.0",
			Author:           "watermelon1024",
			GitHubRepository: "https://github.com/watermelon1024/cpa-codex-retry",
			ConfigFields:     configFields(),
		},
		Capabilities: cliproxy.Capabilities{
			ModelRouter:           true,
			Executor:              true,
			ExecutorModelScope:    "static",
			ExecutorInputFormats:  formats,
			ExecutorOutputFormats: formats,
		},
	}
}

func configFields() []cliproxy.ConfigField {
	return []cliproxy.ConfigField{
		{Name: "enabled", Type: "boolean", Description: "Enable or disable the retry gateway plugin."},
		{Name: "source_formats", Type: "array", Description: "Source protocol formats to guard."},
		{Name: "models", Type: "array", Description: "Exact model names to guard. Empty means all routed models."},
		{Name: "model_prefixes", Type: "array", Description: "Model name prefixes to guard."},
		{Name: "intercept_rule_mode", Type: "enum", EnumValues: []string{config.RuleReasoningTokens, config.RuleFinalOnlyHighXHigh}, Description: "Suspicious response rule."},
		{Name: "reasoning_match_mode", Type: "enum", EnumValues: []string{config.MatchFormula518NMinus2, config.MatchManual}, Description: "Reasoning token matcher."},
		{Name: "reasoning_equals", Type: "array", Description: "Manual reasoning token values when reasoning_match_mode=manual."},
		{Name: "guard_retry_attempts", Type: "integer", Description: "Maximum internal retries after a guard match."},
		{Name: "stream_action", Type: "enum", EnumValues: []string{config.StreamActionContinuation, config.StreamActionStrict502}, Description: "Streaming action after a guard match."},
		{Name: "retry_upstream_capacity_errors", Type: "boolean", Description: "Retry the known upstream capacity error internally."},
	}
}

type hostCallError struct {
	message string
	status  int
}

func (e hostCallError) Error() string {
	return e.message
}

func (e hostCallError) StatusCode() int {
	return e.status
}

func decodeHostEnvelope(method string, code C.int, raw []byte) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("host callback %s returned no response, code=%d", method, int(code))
	}
	var env cliproxy.Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if env.OK {
		return append(json.RawMessage(nil), env.Result...), nil
	}
	return nil, hostEnvelopeError(method, env)
}

func hostEnvelopeError(method string, env cliproxy.Envelope) error {
	if env.Error == nil {
		return fmt.Errorf("host callback %s failed", method)
	}
	return hostCallError{message: env.Error.Message, status: env.Error.HTTPStatus}
}

func callHost(method string, payload any) (json.RawMessage, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	cMethod := C.CString(method)
	defer C.free(unsafe.Pointer(cMethod))
	var response C.cliproxy_buffer
	requestPtr := cBytes(rawPayload)
	if requestPtr != nil {
		defer C.free(unsafe.Pointer(requestPtr))
	}
	code := C.call_host_api(cMethod, requestPtr, C.size_t(len(rawPayload)), &response)
	rawResponse := hostResponseBytes(response)
	return decodeHostEnvelope(method, code, rawResponse)
}

func cBytes(raw []byte) *C.uint8_t {
	if len(raw) == 0 {
		return nil
	}
	ptr := C.CBytes(raw)
	if ptr == nil {
		return nil
	}
	return (*C.uint8_t)(ptr)
}

func hostResponseBytes(response C.cliproxy_buffer) []byte {
	if response.ptr == nil {
		return nil
	}
	var raw []byte
	if response.len > 0 {
		raw = C.GoBytes(response.ptr, C.int(response.len))
	}
	C.free_host_buffer(response.ptr, response.len)
	return raw
}

type cHostClient struct{}

func (cHostClient) ExecuteModel(_ context.Context, req cliproxy.RPCHostModelExecutionRequest) (cliproxy.HostModelExecutionResponse, error) {
	var resp cliproxy.HostModelExecutionResponse
	err := callHostJSON(cliproxy.MethodHostModelExecute, req, &resp)
	return resp, err
}

func (cHostClient) ExecuteModelStream(_ context.Context, req cliproxy.RPCHostModelExecutionRequest) (cliproxy.HostModelStreamResponse, error) {
	var resp cliproxy.HostModelStreamResponse
	err := callHostJSON(cliproxy.MethodHostModelStream, req, &resp)
	return resp, err
}

func (cHostClient) ReadModelStream(_ context.Context, streamID string) (cliproxy.HostModelStreamReadResponse, error) {
	var resp cliproxy.HostModelStreamReadResponse
	err := callHostJSON(cliproxy.MethodHostModelStreamRead, cliproxy.HostModelStreamReadRequest{StreamID: streamID}, &resp)
	return resp, err
}

func (cHostClient) CloseModelStream(_ context.Context, streamID string) error {
	_, err := callHost(cliproxy.MethodHostModelStreamClose, cliproxy.HostModelStreamCloseRequest{StreamID: streamID})
	return err
}

func (cHostClient) Log(_ context.Context, level, message string, fields map[string]any) {
	_, _ = callHost(cliproxy.MethodHostLog, cliproxy.HostLogRequest{
		Level:   level,
		Message: message,
		Fields:  fields,
	})
}

func callHostJSON(method string, payload any, out any) error {
	raw, err := callHost(method, payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}
