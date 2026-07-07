package guard

import (
	"encoding/json"
	"strings"
)

type SSEParser struct {
	buffer string
}

func (p *SSEParser) Push(chunk []byte, current Inspection) (Inspection, bool) {
	normalized := strings.ReplaceAll(string(chunk), "\r\n", "\n")
	if payload, ok := parseRawJSONPayload(normalized); ok {
		return applyStreamPayload(payload, current)
	}
	p.buffer += normalized
	blocks := strings.Split(p.buffer, "\n\n")
	p.buffer = blocks[len(blocks)-1]
	matchedEarly := false
	for _, block := range blocks[:len(blocks)-1] {
		payload, ok := parseSSEPayload(block)
		if !ok {
			continue
		}
		var found bool
		current, found = applyStreamPayload(payload, current)
		if found {
			matchedEarly = true
		}
	}
	return current, matchedEarly
}

func (p *SSEParser) Flush(current Inspection) Inspection {
	if strings.TrimSpace(p.buffer) == "" {
		return current
	}
	payload, ok := parseSSEPayload(p.buffer)
	p.buffer = ""
	if !ok {
		return current
	}
	current, _ = applyStreamPayload(payload, current)
	return current
}

func applyStreamPayload(payload map[string]any, current Inspection) (Inspection, bool) {
	ApplyStructure(payload, &current.Structure, true)
	reasoning, source := ExtractReasoningTokensWithSource(payload)
	if reasoning == nil {
		return current, false
	}
	current.ReasoningTokens = reasoning
	current.ReasoningSource = source
	return current, true
}

func parseSSEPayload(block string) (map[string]any, bool) {
	if payload, ok := parseRawJSONPayload(block); ok {
		return payload, true
	}
	var dataLines []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(line[5:]))
		}
	}
	if len(dataLines) == 0 {
		return nil, false
	}
	payloadText := strings.Join(dataLines, "\n")
	if payloadText == "[DONE]" {
		return nil, false
	}
	var payload map[string]any
	if json.Unmarshal([]byte(payloadText), &payload) != nil {
		return nil, false
	}
	return payload, true
}

func parseRawJSONPayload(raw string) (map[string]any, bool) {
	payloadText := strings.TrimSpace(raw)
	if payloadText == "" || payloadText == "[DONE]" || !strings.HasPrefix(payloadText, "{") {
		return nil, false
	}
	var payload map[string]any
	if json.Unmarshal([]byte(payloadText), &payload) != nil {
		return nil, false
	}
	return payload, true
}
