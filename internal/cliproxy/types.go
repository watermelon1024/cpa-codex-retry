package cliproxy

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const (
	ABIVersion    uint32 = 1
	SchemaVersion uint32 = 1
	PluginID             = "codex-retry-gateway"
)

const (
	MethodPluginRegister        = "plugin.register"
	MethodPluginReconfigure     = "plugin.reconfigure"
	MethodModelRoute            = "model.route"
	MethodExecutorIdentifier    = "executor.identifier"
	MethodExecutorExecute       = "executor.execute"
	MethodExecutorExecuteStream = "executor.execute_stream"
	MethodExecutorCountTokens   = "executor.count_tokens"
	MethodExecutorHTTPRequest   = "executor.http_request"
	MethodManagementRegister    = "management.register"
	MethodManagementHandle      = "management.handle"
	MethodHostModelExecute      = "host.model.execute"
	MethodHostModelStream       = "host.model.execute_stream"
	MethodHostModelStreamRead   = "host.model.stream_read"
	MethodHostModelStreamClose  = "host.model.stream_close"
	MethodHostLog               = "host.log"
)

type Envelope struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *Error          `json:"error,omitempty"`
}

type Error struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Retryable  bool   `json:"retryable,omitempty"`
	HTTPStatus int    `json:"http_status,omitempty"`
}

type LifecycleRequest struct {
	ConfigYAML    []byte `json:"config_yaml"`
	SchemaVersion uint32 `json:"schema_version"`
}

type Metadata struct {
	Name             string        `json:"Name"`
	Version          string        `json:"Version"`
	Author           string        `json:"Author"`
	GitHubRepository string        `json:"GitHubRepository"`
	Logo             string        `json:"Logo,omitempty"`
	ConfigFields     []ConfigField `json:"ConfigFields"`
}

type ConfigField struct {
	Name        string   `json:"Name"`
	Type        string   `json:"Type"`
	EnumValues  []string `json:"EnumValues,omitempty"`
	Description string   `json:"Description"`
}

type Registration struct {
	SchemaVersion uint32       `json:"schema_version"`
	Metadata      Metadata     `json:"metadata"`
	Capabilities  Capabilities `json:"capabilities"`
}

type Capabilities struct {
	ModelRouter           bool     `json:"model_router"`
	Executor              bool     `json:"executor"`
	ExecutorModelScope    string   `json:"executor_model_scope"`
	ExecutorInputFormats  []string `json:"executor_input_formats,omitempty"`
	ExecutorOutputFormats []string `json:"executor_output_formats,omitempty"`
	ManagementAPI         bool     `json:"management_api"`
}

type ManagementRegistrationResponse struct {
	Resources []ResourceRoute `json:"resources,omitempty"`
}

type ResourceRoute struct {
	Path        string `json:"Path"`
	Menu        string `json:"Menu"`
	Description string `json:"Description"`
}

type ManagementRequest struct {
	Method         string      `json:"Method"`
	Path           string      `json:"Path"`
	Headers        http.Header `json:"Headers,omitempty"`
	Query          url.Values  `json:"Query,omitempty"`
	Body           []byte      `json:"Body,omitempty"`
	HostCallbackID string      `json:"host_callback_id,omitempty"`
}

type ManagementResponse struct {
	StatusCode int         `json:"StatusCode"`
	Headers    http.Header `json:"Headers,omitempty"`
	Body       []byte      `json:"Body,omitempty"`
}

type ModelRouteRequest struct {
	PluginID           string      `json:"PluginID,omitempty"`
	SourceFormat       string      `json:"SourceFormat"`
	RequestedModel     string      `json:"RequestedModel"`
	Stream             bool        `json:"Stream"`
	Headers            http.Header `json:"Headers,omitempty"`
	Query              url.Values  `json:"Query,omitempty"`
	Body               []byte      `json:"Body,omitempty"`
	AvailableProviders []string    `json:"AvailableProviders,omitempty"`
}

type ModelRouteResponse struct {
	Handled     bool   `json:"Handled"`
	TargetKind  string `json:"TargetKind,omitempty"`
	Target      string `json:"Target,omitempty"`
	TargetModel string `json:"TargetModel,omitempty"`
	Reason      string `json:"Reason,omitempty"`
}

type RPCExecutorRequest struct {
	ExecutorRequest
	StreamID       string `json:"stream_id,omitempty"`
	HostCallbackID string `json:"host_callback_id,omitempty"`
}

type ExecutorRequest struct {
	AuthID          string            `json:"AuthID,omitempty"`
	AuthProvider    string            `json:"AuthProvider,omitempty"`
	Model           string            `json:"Model"`
	Format          string            `json:"Format"`
	Stream          bool              `json:"Stream"`
	Alt             string            `json:"Alt,omitempty"`
	Headers         http.Header       `json:"Headers,omitempty"`
	Query           url.Values        `json:"Query,omitempty"`
	OriginalRequest []byte            `json:"OriginalRequest,omitempty"`
	SourceFormat    string            `json:"SourceFormat,omitempty"`
	Payload         []byte            `json:"Payload,omitempty"`
	Metadata        map[string]any    `json:"Metadata,omitempty"`
	AuthMetadata    map[string]any    `json:"AuthMetadata,omitempty"`
	AuthAttributes  map[string]string `json:"AuthAttributes,omitempty"`
}

type ExecutorResponse struct {
	Payload []byte      `json:"Payload,omitempty"`
	Headers http.Header `json:"Headers,omitempty"`
}

type ExecutorStreamResponse struct {
	Headers http.Header           `json:"headers,omitempty"`
	Chunks  []ExecutorStreamChunk `json:"chunks,omitempty"`
}

type ExecutorStreamChunk struct {
	Payload []byte `json:"Payload,omitempty"`
	Error   string `json:"Err,omitempty"`
}

type HostModelExecutionRequest struct {
	EntryProtocol string      `json:"entry_protocol"`
	ExitProtocol  string      `json:"exit_protocol"`
	Model         string      `json:"model"`
	Stream        bool        `json:"stream"`
	Body          []byte      `json:"body"`
	Headers       http.Header `json:"headers,omitempty"`
	Query         url.Values  `json:"query,omitempty"`
	Alt           string      `json:"alt,omitempty"`
}

type RPCHostModelExecutionRequest struct {
	HostModelExecutionRequest
	HostCallbackID string `json:"host_callback_id,omitempty"`
}

type HostModelExecutionResponse struct {
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers,omitempty"`
	Body       []byte      `json:"body,omitempty"`
}

type HostModelStreamResponse struct {
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers,omitempty"`
	StreamID   string      `json:"stream_id"`
}

type HostModelStreamReadRequest struct {
	StreamID string `json:"stream_id"`
}

type HostModelStreamReadResponse struct {
	Payload []byte `json:"payload,omitempty"`
	Error   string `json:"error,omitempty"`
	Done    bool   `json:"done"`
}

type HostModelStreamCloseRequest struct {
	StreamID string `json:"stream_id"`
}

type HostLogRequest struct {
	Level          string         `json:"level"`
	Message        string         `json:"message"`
	Fields         map[string]any `json:"fields,omitempty"`
	HostCallbackID string         `json:"host_callback_id,omitempty"`
}
