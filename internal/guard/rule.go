package guard

import (
	"fmt"
	"strings"

	"github.com/watermelon1024/cpa-codex-retry/internal/config"
)

func Match(cfg config.Config, inspection Inspection, effort string) Decision {
	reasoning := inspection.ReasoningTokens
	if inspection.RequestKind == RequestKindContextCompaction && intValue(reasoning) == 0 {
		return Decision{Exempt: true, Mode: cfg.InterceptRuleMode, ReasoningTokens: reasoning, ReasoningSource: inspection.ReasoningSource, Reason: "context_compaction"}
	}
	if cfg.InterceptRuleMode == RuleFinalOnlyHighXHigh {
		return matchFinalOnly(cfg, inspection, effort)
	}
	return matchReasoningTokens(cfg, inspection)
}

func matchReasoningTokens(cfg config.Config, inspection Inspection) Decision {
	reasoning := inspection.ReasoningTokens
	if reasoning == nil || !reasoningMatched(cfg, *reasoning) {
		return Decision{Mode: RuleReasoningTokens, ReasoningTokens: reasoning, ReasoningSource: inspection.ReasoningSource}
	}
	value := *reasoning
	return Decision{
		Matched:          true,
		Mode:             RuleReasoningTokens,
		ReasoningTokens:  &value,
		ReasoningSource:  inspection.ReasoningSource,
		BlockedReasoning: &value,
		Reason:           fmt.Sprintf("reasoning_tokens=%d", value),
	}
}

func matchFinalOnly(cfg config.Config, inspection Inspection, effort string) Decision {
	reasoning := intValue(inspection.ReasoningTokens)
	effort = strings.ToLower(strings.TrimSpace(effort))
	if !inspection.Structure.FinalAnswerOnly() || reasoning == 0 || !highEffort(effort) {
		return Decision{Mode: cfg.InterceptRuleMode, ReasoningTokens: inspection.ReasoningTokens, ReasoningSource: inspection.ReasoningSource}
	}
	return Decision{
		Matched:          true,
		Mode:             cfg.InterceptRuleMode,
		ReasoningTokens:  inspection.ReasoningTokens,
		ReasoningSource:  inspection.ReasoningSource,
		BlockedReasoning: inspection.ReasoningTokens,
		Reason:           "final_answer_only_high_xhigh",
	}
}

func reasoningMatched(cfg config.Config, reasoning int) bool {
	if cfg.ReasoningMatchMode == config.MatchFormula518NMinus2 {
		return reasoning >= 516 && (reasoning+2)%518 == 0
	}
	for _, value := range cfg.ReasoningEquals {
		if value == reasoning {
			return true
		}
	}
	return false
}

func highEffort(effort string) bool {
	return effort == "high" || effort == "xhigh"
}

func intValue(value *int) int {
	if value == nil {
		return -1
	}
	return *value
}
