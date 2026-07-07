# Codex Retry Gateway CLIProxyAPI 插件

語言： [English](README.md) | **繁體中文** | [简体中文](README.zh-CN.md)

這個專案目前保留為 CLIProxyAPI 插件實驗。

## 目前狀態

> [!CAUTION]
> 這個插件無法有效完成原本設計的 reasoning token guard。
>
> CLIProxyAPI 會在模型執行完成後，才把 reasoning token 寫進 usage record。插件 executor callback 只能拿到 response body 或 stream chunk，而這些 payload 不一定會帶最終 reasoning token 數量。因此插件無法在請求完成前可靠取得 `reasoning_tokens`，也就無法可靠地重試或阻擋可疑回應。

## 已實作範圍

- CLIProxyAPI 插件 executor 註冊與管理面板。
- 在 returned payload 仍包含 usage 時，檢查 response body。
- 對偵測到的可疑 reasoning token 模式做重試或阻擋。
- 針對部分 Responses 串流 payload 做恢復處理。

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
