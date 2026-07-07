# Codex Retry Gateway CLIProxyAPI 插件

语言： [English](README.md) | [繁體中文](README.zh-TW.md) | **简体中文**

这个插件给 CLIProxyAPI 加上一层 reasoning response guard。

它会把指定的 Codex/OpenAI 请求路由到 CLIProxyAPI 插件 executor，通过 CLIProxyAPI 原本的 upstream model execution callback 发出请求，检查上游响应，并在命中已知可疑 reasoning 模式时重试或阻挡。

目的很简单：让 client 照常连 CLIProxyAPI，同时多一层保护，用来处理 reasoning token 异常和可恢复的 upstream capacity 错误。

## 功能

- 保护 Codex/OpenAI 兼容请求格式：`codex`、`openai-response`、`openai`。
- 非流式响应检测到可疑 `reasoning_tokens` 时自动重试。
- 流式响应会先缓冲并检查，再转发给下游。
- 支持 Responses 流式续写恢复。
- 默认检测 `reasoning_tokens = 518*n - 2`，例如 `516 / 1034 / 1552 / 2070 ...`。
- 支持手动 `reasoning_equals` 匹配。
- 支持实验规则 `final_answer_only_high_xhigh`。
- 明确的 `context_compaction` 响应在 `reasoning_tokens=0` 时豁免。
- 针对已知 upstream capacity 错误做内部重试。
- 续写 replay 会清理 `previous_response_id`、replayed reasoning items、`reasoning.encrypted_content` 和 `encrypted_content`。

## 构建

```bash
make build
```

产物会输出到：

```text
dist/<goos>/<goarch>/codex-retry-gateway.<dylib|so|dll>
```

macOS 手动构建示例：

```bash
go build -buildmode=c-shared -o plugins/darwin/$(go env GOARCH)/codex-retry-gateway.dylib ./cmd/codex-retry-gateway-plugin
```

## CLIProxyAPI 配置

把构建好的动态库放进 CLIProxyAPI 插件目录，然后启用：

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

## 配置字段

- `enabled`：开启或关闭插件。
- `source_formats`：插件要保护的请求协议格式。
- `models`：要保护的完整模型名称。留空代表所有被路由到的模型。
- `model_prefixes`：要保护的模型名称前缀。
- `intercept_rule_mode`：`reasoning_tokens` 或 `final_answer_only_high_xhigh`。
- `reasoning_match_mode`：`formula_518n_minus_2` 或 `manual`。
- `reasoning_equals`：当 `reasoning_match_mode=manual` 时使用的 reasoning token 数值。
- `guard_retry_attempts`：命中规则后最多内部重试几次。
- `retry_upstream_capacity_errors`：针对已知 upstream capacity 错误做内部重试。
- `stream_action`：`continuation_recovery` 或 `strict_502`。
- `intercept_streaming`：启用流式响应 guard。
- `intercept_non_streaming`：启用非流式响应 guard。
- `non_stream_status_code`：插件阻挡响应时使用的 HTTP 错误状态码。

## 运行注意事项

流式响应会被缓冲，这样插件才能在 CLIProxyAPI 送出下游 response headers 前完成检查。使用 `continuation_recovery` 可以自动做 Responses replay；使用 `strict_502` 则会在重试后仍命中 guard 时返回错误。

## Credit

本项目参考原始 `codex-retry-gateway` 项目：<https://github.com/nonononull/codex-retry-gateway>。
