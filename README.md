# Codex Retry Gateway CLIProxyAPI Plugin

Languages: **English** | [繁體中文](README.zh-TW.md) | [简体中文](README.zh-CN.md)

This project is currently kept as a CLIProxyAPI plugin experiment.

## Current Status

> [!CAUTION]
> This plugin is not effective for its intended reasoning-token guard.
>
> CLIProxyAPI records reasoning tokens in usage records after model execution. The plugin executor callback only receives the response body or stream chunks, and those payloads do not reliably include the final reasoning token count. Because the plugin cannot reliably read `reasoning_tokens` before the request completes, it cannot reliably retry or block suspicious responses.

## Implemented Surface

- CLIProxyAPI plugin executor registration and management panel.
- Response/body inspection for cases where usage is still present in the returned payload.
- Retry and block logic for detected suspicious reasoning-token patterns.
- Recovery logic for selected streaming Responses payloads.

## Build

```bash
make build
```

The output is written to:

```text
dist/<goos>/<goarch>/codex-retry-gateway.<dylib|so|dll>
```

Manual macOS build example:

```bash
go build -buildmode=c-shared -o plugins/darwin/$(go env GOARCH)/codex-retry-gateway.dylib ./cmd/codex-retry-gateway-plugin
```

## CLIProxyAPI Config

Copy the built library into your CLIProxyAPI plugin directory, then enable it:

```yaml
plugins:
  enabled: true
  dir: "plugins"
  configs:
    codex-retry-gateway:
      enabled: true
      priority: 1
      source_formats: ["codex", "openai-response", "openai"]
      intercept_rule_mode: "reasoning_tokens"
      reasoning_match_mode: "formula_518n_minus_2"
      reasoning_equals: [516, 1034, 1552]
      guard_retry_attempts: 5
      retry_upstream_capacity_errors: true
      stream_action: "continuation_recovery"
      intercept_streaming: true
      intercept_non_streaming: true
      non_stream_status_code: 502
```

## Config Fields

- `enabled`: turns the plugin on or off.
- `source_formats`: request protocol formats guarded by the plugin.
- `models`: exact model names to guard. Empty means all routed models.
- `model_prefixes`: model name prefixes to guard.
- `intercept_rule_mode`: `reasoning_tokens` or `final_answer_only_high_xhigh`.
- `reasoning_match_mode`: `formula_518n_minus_2` or `manual`.
- `reasoning_equals`: manual reasoning token values when `reasoning_match_mode=manual`.
- `guard_retry_attempts`: maximum internal retry count after a guard match.
- `retry_upstream_capacity_errors`: retries the known upstream capacity error internally.
- `stream_action`: `continuation_recovery` or `strict_502`.
- `intercept_streaming`: enables guard behavior for streaming responses.
- `intercept_non_streaming`: enables guard behavior for non-streaming responses.
- `non_stream_status_code`: HTTP error status used when the plugin blocks a response.

## Runtime Notes

Streaming responses are buffered so the plugin can inspect the result before CLIProxyAPI commits downstream response headers. Use `continuation_recovery` for automatic Responses replay, or `strict_502` to return an error when a guarded stream still matches after retries.

## Credits

This project references the original `codex-retry-gateway` project: <https://github.com/nonononull/codex-retry-gateway>.
