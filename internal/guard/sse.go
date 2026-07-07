package guard

import (
	"encoding/json"
	"strings"
)

type SSEParser struct {
	buffer string
}

func (p *SSEParser) Push(chunk []byte, current Inspection) (Inspection, bool) {
	p.buffer += strings.ReplaceAll(string(chunk), "\r\n", "\n")
	blocks := strings.Split(p.buffer, "\n\n")
	p.buffer = blocks[len(blocks)-1]
	matchedEarly := false
	for _, block := range blocks[:len(blocks)-1] {
		payload, ok := parseSSEPayload(block)
		if !ok {
			continue
		}
		ApplyStructure(payload, &current.Structure, true)
		if reasoning := ExtractReasoningTokens(payload); reasoning != nil {
			current.ReasoningTokens = reasoning
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
	ApplyStructure(payload, &current.Structure, true)
	if reasoning := ExtractReasoningTokens(payload); reasoning != nil {
		current.ReasoningTokens = reasoning
	}
	return current
}

func parseSSEPayload(block string) (map[string]any, bool) {
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
