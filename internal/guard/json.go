package guard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

var reasoningPointers = [][]string{
	{"usage", "output_tokens_details", "reasoning_tokens"},
	{"usage", "completion_tokens_details", "reasoning_tokens"},
	{"usage", "reasoning_tokens"},
	{"usage", "total_thought_tokens"},
	{"usage", "thoughtsTokenCount"},
	{"response", "usage", "output_tokens_details", "reasoning_tokens"},
	{"response", "usage", "completion_tokens_details", "reasoning_tokens"},
	{"response", "usage", "reasoning_tokens"},
	{"response", "usage", "total_thought_tokens"},
	{"response", "usage", "thoughtsTokenCount"},
	{"interaction", "usage", "reasoning_tokens"},
	{"interaction", "usage", "total_thought_tokens"},
	{"interaction", "usage", "thoughtsTokenCount"},
	{"interaction", "total_usage", "reasoning_tokens"},
	{"interaction", "total_usage", "total_thought_tokens"},
	{"interaction", "metadata", "total_usage", "reasoning_tokens"},
	{"interaction", "metadata", "total_usage", "total_thought_tokens"},
	{"metadata", "usage", "reasoning_tokens"},
	{"metadata", "usage", "total_thought_tokens"},
	{"metadata", "total_usage", "reasoning_tokens"},
	{"metadata", "total_usage", "total_thought_tokens"},
	{"total_usage", "reasoning_tokens"},
	{"total_usage", "total_thought_tokens"},
	{"usageMetadata", "thoughtsTokenCount"},
	{"usage_metadata", "thoughtsTokenCount"},
	{"usage_metadata", "thoughts_token_count"},
	{"usage_metadata", "total_thought_tokens"},
	{"tokens", "reasoning_tokens"},
	{"detail", "reasoning_tokens"},
	{"reasoning_tokens"},
	{"total_thought_tokens"},
	{"thoughtsTokenCount"},
	{"thoughts_token_count"},
}

func InspectJSON(body []byte, headers http.Header, requestBody []byte) Inspection {
	payload := decodeObject(body)
	requestPayload := decodeObject(requestBody)
	structure := Structure{}
	if payload != nil {
		ApplyStructure(payload, &structure, false)
	}
	reasoning, source := ExtractReasoningTokensWithSource(payload)
	return Inspection{
		ReasoningTokens: reasoning,
		ReasoningSource: source,
		Structure:       structure,
		RequestKind:     DetectRequestKind(headers, requestPayload),
	}
}

func ExtractReasoningTokens(payload map[string]any) *int {
	value, _ := ExtractReasoningTokensWithSource(payload)
	return value
}

func ExtractReasoningTokensWithSource(payload map[string]any) (*int, string) {
	for _, pointer := range reasoningPointers {
		if value, ok := intFromJSONValue(nested(payload, pointer)); ok {
			return &value, strings.Join(pointer, ".")
		}
	}
	return nil, ""
}

func intFromJSONValue(value any) (int, bool) {
	switch typed := value.(type) {
	case float64:
		intValue := int(typed)
		return intValue, float64(intValue) == typed
	case json.Number:
		intValue, err := strconv.Atoi(typed.String())
		return intValue, err == nil
	case string:
		intValue, err := strconv.Atoi(strings.TrimSpace(typed))
		return intValue, err == nil
	case int:
		return typed, true
	case int64:
		return int(typed), int64(int(typed)) == typed
	default:
		return 0, false
	}
}

func ApplyStructure(payload map[string]any, structure *Structure, fromStream bool) {
	if payload == nil || structure == nil {
		return
	}
	inspectPayloadType(payload, structure)
	inspectChoices(payload["choices"], structure)
	inspectOutputText(payload, structure)
	inspectOutputCollections(payload, structure)
	if item, ok := payload["item"].(map[string]any); ok {
		inspectOutputItem(item, structure)
	}
	_ = fromStream
}

func DetectRequestKind(headers http.Header, requestPayload map[string]any) string {
	headerSignals := []string{
		headers.Get("x-codex-request-kind"),
		headers.Get("x-codex-purpose"),
		headers.Get("x-codex-turn-metadata"),
	}
	if hasCompactionMarker(strings.Join(headerSignals, " ")) {
		return RequestKindContextCompaction
	}
	if requestPayload == nil {
		return RequestKindNormal
	}
	return requestKindFromPayload(requestPayload)
}

func decodeObject(raw []byte) map[string]any {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	var payload map[string]any
	if json.Unmarshal(raw, &payload) != nil {
		return nil
	}
	return payload
}

func nested(root map[string]any, path []string) any {
	var current any = root
	for _, part := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = object[part]
	}
	return current
}

func inspectPayloadType(payload map[string]any, structure *Structure) {
	eventType := lowerString(payload["type"])
	if strings.Contains(eventType, "commentary") {
		structure.HasCommentary = true
	}
	if strings.Contains(eventType, "tool_call") || strings.Contains(eventType, "function_call") {
		structure.HasToolCall = true
	}
	if strings.Contains(eventType, "output_text.delta") || strings.Contains(eventType, "message.delta") {
		markVisibleIfText(payload["delta"], structure)
		markVisibleIfText(payload["text"], structure)
		markVisibleIfText(payload["content"], structure)
	}
}

func inspectChoices(value any, structure *Structure) {
	choices, ok := value.([]any)
	if !ok {
		return
	}
	for _, rawChoice := range choices {
		choice, okChoice := rawChoice.(map[string]any)
		if !okChoice {
			continue
		}
		inspectChoiceText(choice["delta"], structure)
		inspectChoiceText(choice["message"], structure)
	}
}

func inspectChoiceText(value any, structure *Structure) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}
	markVisibleIfText(object["content"], structure)
}

func inspectOutputText(payload map[string]any, structure *Structure) {
	markVisibleIfText(payload["output_text"], structure)
	markVisibleIfText(payload["text"], structure)
}

func inspectOutputCollections(payload map[string]any, structure *Structure) {
	for _, value := range []any{payload["output"], nested(payload, []string{"response", "output"})} {
		items, ok := value.([]any)
		if !ok {
			continue
		}
		for _, rawItem := range items {
			if item, okItem := rawItem.(map[string]any); okItem {
				inspectOutputItem(item, structure)
			}
		}
	}
}

func inspectOutputItem(item map[string]any, structure *Structure) {
	itemType := lowerString(item["type"])
	if strings.Contains(itemType, "reasoning") {
		structure.HasReasoningItem = true
	}
	if strings.Contains(itemType, "commentary") {
		structure.HasCommentary = true
	}
	if strings.Contains(itemType, "tool") || strings.Contains(itemType, "function_call") {
		structure.HasToolCall = true
	}
	markVisibleIfText(item["text"], structure)
	markVisibleIfText(item["output_text"], structure)
	inspectContentEntries(item["content"], structure)
}

func inspectContentEntries(value any, structure *Structure) {
	entries, ok := value.([]any)
	if !ok {
		return
	}
	for _, rawEntry := range entries {
		if entry, okEntry := rawEntry.(map[string]any); okEntry {
			inspectContentEntry(entry, structure)
		}
	}
}

func inspectContentEntry(entry map[string]any, structure *Structure) {
	contentType := lowerString(entry["type"])
	if strings.Contains(contentType, "commentary") {
		structure.HasCommentary = true
	}
	if strings.Contains(contentType, "tool_call") || strings.Contains(contentType, "function_call") {
		structure.HasToolCall = true
	}
	markVisibleIfText(entry["text"], structure)
	markVisibleIfText(entry["output_text"], structure)
	markVisibleIfText(entry["content"], structure)
}

func markVisibleIfText(value any, structure *Structure) {
	text, ok := value.(string)
	if ok && strings.TrimSpace(text) != "" {
		structure.HasFinalAnswer = true
		structure.HasOutputText = true
	}
}

func requestKindFromPayload(payload map[string]any) string {
	for _, key := range []string{"metadata", "codex_request_kind", "request_kind", "purpose"} {
		if hasCompactionMarker(stringifySignal(payload[key])) {
			return RequestKindContextCompaction
		}
	}
	return RequestKindNormal
}

func stringifySignal(value any) string {
	if value == nil {
		return ""
	}
	if raw, err := json.Marshal(value); err == nil {
		return string(raw)
	}
	return ""
}

func hasCompactionMarker(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.Contains(normalized, "remote_compaction") ||
		strings.Contains(normalized, "context_compaction")
}

func lowerString(value any) string {
	text, _ := value.(string)
	return strings.ToLower(strings.TrimSpace(text))
}
