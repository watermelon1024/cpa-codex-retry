package guard

const (
	RequestKindNormal            = "normal"
	RequestKindContextCompaction = "context_compaction"
	RuleReasoningTokens          = "reasoning_tokens"
	RuleFinalOnlyHighXHigh       = "final_answer_only_high_xhigh"
)

type Structure struct {
	HasCommentary    bool
	HasFinalAnswer   bool
	HasToolCall      bool
	HasOutputText    bool
	HasReasoningItem bool
}

type Decision struct {
	Matched          bool
	Exempt           bool
	Mode             string
	BlockedReasoning *int
	Reason           string
}

type Inspection struct {
	ReasoningTokens *int
	Structure       Structure
	RequestKind     string
}

func (s Structure) FinalAnswerOnly() bool {
	return s.HasFinalAnswer && !s.HasCommentary && !s.HasToolCall && !s.HasReasoningItem
}
