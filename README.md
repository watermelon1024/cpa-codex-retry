# Codex Retry Gateway CLIProxyAPI Plugin

Languages: **English** | [繁體中文](README.zh-TW.md) | [简体中文](README.zh-CN.md)

This plugin adds a reasoning-response guard to CLIProxyAPI.

It routes selected Codex/OpenAI requests through a CLIProxyAPI plugin executor, sends the request through CLIProxyAPI's normal upstream model execution callbacks, inspects the upstream response, and retries or blocks responses that match known suspicious reasoning patterns.

The goal is simple: keep the client talking to CLIProxyAPI normally, while adding an extra protection layer for reasoning-token anomalies and recoverable upstream capacity failures.

## Features

- Guards Codex/OpenAI compatible request formats: `codex`, `openai-response`, and `openai`.
- Retries non-streaming responses when suspicious `reasoning_tokens` are detected.
- Buffers and inspects streaming responses before forwarding them downstream.
- Supports Responses streaming continuation recovery.
- Detects the default `reasoning_tokens = 518*n - 2` pattern, such as `516 / 1034 / 1552 / 2070 ...`.
- Supports manual `reasoning_equals` matching.
- Supports the experimental `final_answer_only_high_xhigh` rule.
- Exempts explicit `context_compaction` responses when `reasoning_tokens=0`.
- Retries the known upstream capacity error internally.
- Sanitizes continuation replay by removing `previous_response_id`, replayed reasoning items, `reasoning.encrypted_content`, and `encrypted_content`.

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
