package guard

import (
	"bytes"
	"encoding/json"

	"github.com/watermelon1024/cpa-codex-retry/internal/config"
)

const encryptedInclude = "reasoning.encrypted_content"

func BuildContinuationRequest(cfg config.Config, baseBody []byte) ([]byte, error) {
	var body map[string]any
	if err := json.Unmarshal(baseBody, &body); err != nil {
		return nil, err
	}
	delete(body, "previous_response_id")
	updateContinuationInclude(body)
	body["stream"] = true
	body["input"] = continuationInput(cfg, body["input"])
	return json.Marshal(body)
}

func IsResponsesRequest(body []byte) bool {
	var payload map[string]any
	if json.Unmarshal(body, &payload) != nil {
		return false
	}
	_, ok := payload["input"]
	return ok
}

func StripEncryptedContentJSON(raw []byte) []byte {
	var payload any
	if json.Unmarshal(raw, &payload) != nil {
		return append([]byte(nil), raw...)
	}
	stripped := stripEncrypted(payload)
	out, err := json.Marshal(stripped)
	if err != nil {
		return append([]byte(nil), raw...)
	}
	return out
}

func StripEncryptedContentSSE(raw []byte) []byte {
	blocks := bytes.Split(normalizeNewlines(raw), []byte("\n\n"))
	out := make([][]byte, 0, len(blocks))
	for _, block := range blocks {
		out = append(out, stripEncryptedSSEBlock(block))
	}
	return bytes.Join(out, []byte("\n\n"))
}

func updateContinuationInclude(body map[string]any) {
	include, ok := body["include"].([]any)
	if !ok {
		return
	}
	next := make([]any, 0, len(include))
	for _, value := range include {
		if text, okText := value.(string); !okText || text != encryptedInclude {
			next = append(next, value)
		}
	}
	if len(next) == 0 {
		delete(body, "include")
		return
	}
	body["include"] = next
}

func continuationInput(cfg config.Config, rawInput any) []any {
	items := normalizeInputItems(rawInput)
	items = append(items, map[string]any{
		"type":  "message",
		"role":  "assistant",
		"phase": "commentary",
		"content": []any{map[string]any{
			"type": "output_text",
			"text": cfg.ContinuationMarkerText,
		}},
	})
	return items
}

func normalizeInputItems(rawInput any) []any {
	rawItems, ok := rawInput.([]any)
	if !ok && rawInput != nil {
		rawItems = []any{rawInput}
	}
	items := make([]any, 0, len(rawItems))
	for _, rawItem := range rawItems {
		if item := normalizeInputItem(rawItem); item != nil {
			items = append(items, item)
		}
	}
	return items
}

func normalizeInputItem(rawItem any) any {
	if text, ok := rawItem.(string); ok {
		return map[string]any{"type": "message", "role": "user", "content": text}
	}
	item, ok := stripEncrypted(rawItem).(map[string]any)
	if !ok {
		return stripEncrypted(rawItem)
	}
	if itemType, _ := item["type"].(string); itemType == "reasoning" {
		return nil
	}
	return item
}

func stripEncrypted(value any) any {
	switch typed := value.(type) {
	case []any:
		return stripEncryptedSlice(typed)
	case map[string]any:
		return stripEncryptedMap(typed)
	default:
		return value
	}
}

func stripEncryptedSlice(values []any) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, stripEncrypted(value))
	}
	return out
}

func stripEncryptedMap(values map[string]any) map[string]any {
	out := make(map[string]any, len(values))
	for key, value := range values {
		if key != "encrypted_content" {
			out[key] = stripEncrypted(value)
		}
	}
	return out
}

func normalizeNewlines(raw []byte) []byte {
	withoutCRLF := bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	return bytes.ReplaceAll(withoutCRLF, []byte("\r"), []byte("\n"))
}

func stripEncryptedSSEBlock(block []byte) []byte {
	lines := bytes.Split(block, []byte("\n"))
	payload, indexes := sseDataPayload(lines)
	if len(indexes) == 0 || bytes.Equal(bytes.TrimSpace(payload), []byte("[DONE]")) {
		return block
	}
	stripped := StripEncryptedContentJSON(payload)
	lines[indexes[0]] = append([]byte("data: "), stripped...)
	for _, index := range indexes[1:] {
		lines[index] = nil
	}
	return bytes.Join(compactLines(lines), []byte("\n"))
}

func sseDataPayload(lines [][]byte) ([]byte, []int) {
	var payloads [][]byte
	var indexes []int
	for index, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmed, []byte("data:")) {
			payloads = append(payloads, bytes.TrimSpace(trimmed[5:]))
			indexes = append(indexes, index)
		}
	}
	return bytes.Join(payloads, []byte("\n")), indexes
}

func compactLines(lines [][]byte) [][]byte {
	out := make([][]byte, 0, len(lines))
	for _, line := range lines {
		if line != nil {
			out = append(out, line)
		}
	}
	return out
}
