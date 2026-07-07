# Codex Retry Gateway CLIProxyAPI 插件

语言： [English](README.md) | [繁體中文](README.zh-TW.md) | **简体中文**

这个项目目前保留为 CLIProxyAPI 插件实验。

## 当前状态

> [!CAUTION]
> 这个插件无法有效完成原本设计的 reasoning token guard。
>
> CLIProxyAPI 会在模型执行完成后，才把 reasoning token 写进 usage record。插件 executor callback 只能拿到 response body 或 stream chunk，而这些 payload 不一定会带最终 reasoning token 数量。因此插件无法在请求完成前可靠取得 `reasoning_tokens`，也就无法可靠地重试或阻挡可疑响应。

## 已实现范围

- CLIProxyAPI 插件 executor 注册与管理面板。
- 在 returned payload 仍包含 usage 时，检查 response body。
- 对检测到的可疑 reasoning token 模式做重试或阻挡。
- 针对部分 Responses 流式 payload 做恢复处理。

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
