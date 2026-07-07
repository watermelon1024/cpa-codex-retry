# Codex Retry Gateway CLIProxyAPI 插件

語言： [English](README.md) | **繁體中文** | [简体中文](README.zh-CN.md)

這個插件替 CLIProxyAPI 加上一層 reasoning response guard。

它會把指定的 Codex/OpenAI 請求路由到 CLIProxyAPI 插件 executor，透過 CLIProxyAPI 原本的 upstream model execution callback 發出請求，檢查上游回應，並在命中已知可疑 reasoning 模式時重試或阻擋。

目的很單純：讓 client 照常連 CLIProxyAPI，同時多一層保護，用來處理 reasoning token 異常和可恢復的 upstream capacity 錯誤。

## 功能

- 保護 Codex/OpenAI 相容請求格式：`codex`、`openai-response`、`openai`。
- 非串流回應偵測到可疑 `reasoning_tokens` 時自動重試。
- 串流回應會先緩衝並檢查，再轉發給下游。
- 支援 Responses 串流續寫恢復。
- 預設偵測 `reasoning_tokens = 518*n - 2`，例如 `516 / 1034 / 1552 / 2070 ...`。
- 支援手動 `reasoning_equals` 匹配。
- 支援實驗規則 `final_answer_only_high_xhigh`。
- 明確的 `context_compaction` 回應在 `reasoning_tokens=0` 時豁免。
- 針對已知 upstream capacity 錯誤做內部重試。
- 續寫 replay 會清理 `previous_response_id`、replayed reasoning items、`reasoning.encrypted_content` 和 `encrypted_content`。

## 建置

```bash
make build
```

產物會輸出到：

```text
dist/<goos>/<goarch>/codex-retry-gateway.<dylib|so|dll>
```

macOS 手動建置範例：

```bash
go build -buildmode=c-shared -o plugins/darwin/$(go env GOARCH)/codex-retry-gateway.dylib ./cmd/codex-retry-gateway-plugin
```

## CLIProxyAPI 設定

把建置好的動態庫放進 CLIProxyAPI 插件目錄，然後啟用：

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

## 設定欄位

- `enabled`：開啟或關閉插件。
- `source_formats`：插件要保護的請求協議格式。
- `models`：要保護的完整模型名稱。留空代表所有被路由到的模型。
- `model_prefixes`：要保護的模型名稱前綴。
- `intercept_rule_mode`：`reasoning_tokens` 或 `final_answer_only_high_xhigh`。
- `reasoning_match_mode`：`formula_518n_minus_2` 或 `manual`。
- `reasoning_equals`：當 `reasoning_match_mode=manual` 時使用的 reasoning token 數值。
- `guard_retry_attempts`：命中規則後最多內部重試幾次。
- `retry_upstream_capacity_errors`：針對已知 upstream capacity 錯誤做內部重試。
- `stream_action`：`continuation_recovery` 或 `strict_502`。
- `intercept_streaming`：啟用串流回應 guard。
- `intercept_non_streaming`：啟用非串流回應 guard。
- `non_stream_status_code`：插件阻擋回應時使用的 HTTP 錯誤狀態碼。

## 執行注意事項

串流回應會被緩衝，這樣插件才能在 CLIProxyAPI 送出下游 response headers 前完成檢查。使用 `continuation_recovery` 可以自動做 Responses replay；使用 `strict_502` 則會在重試後仍命中 guard 時回傳錯誤。

## Credit

本專案參考原始 `codex-retry-gateway` 專案：<https://github.com/nonononull/codex-retry-gateway>。
