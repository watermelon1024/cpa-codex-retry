package config

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	RuleReasoningTokens      = "reasoning_tokens"
	RuleFinalOnlyHighXHigh   = "final_answer_only_high_xhigh"
	MatchManual              = "manual"
	MatchFormula518NMinus2   = "formula_518n_minus_2"
	StreamActionStrict502    = "strict_502"
	StreamActionContinuation = "continuation_recovery"
)

type Config struct {
	Enabled                     bool
	SourceFormats               []string
	Models                      []string
	ModelPrefixes               []string
	InterceptRuleMode           string
	ReasoningMatchMode          string
	ReasoningEquals             []int
	ContinuationMarkerText      string
	InterceptStreaming          bool
	InterceptNonStreaming       bool
	NonStreamStatusCode         int
	GuardRetryAttempts          int
	RetryUpstreamCapacityErrors bool
	StreamAction                string
	LogMatch                    bool
	StripEncryptedContent       bool
}

type rawConfig struct {
	Enabled                     *bool    `yaml:"enabled"`
	SourceFormats               []string `yaml:"source_formats"`
	Models                      []string `yaml:"models"`
	ModelPrefixes               []string `yaml:"model_prefixes"`
	InterceptRuleMode           string   `yaml:"intercept_rule_mode"`
	ReasoningMatchMode          string   `yaml:"reasoning_match_mode"`
	ReasoningEquals             intList  `yaml:"reasoning_equals"`
	ContinuationMarkerText      string   `yaml:"continuation_marker_text"`
	InterceptStreaming          *bool    `yaml:"intercept_streaming"`
	InterceptNonStreaming       *bool    `yaml:"intercept_non_streaming"`
	NonStreamStatusCode         int      `yaml:"non_stream_status_code"`
	GuardRetryAttempts          *int     `yaml:"guard_retry_attempts"`
	RetryUpstreamCapacityErrors *bool    `yaml:"retry_upstream_capacity_errors"`
	StreamAction                string   `yaml:"stream_action"`
	LogMatch                    *bool    `yaml:"log_match"`
	StripEncryptedContent       *bool    `yaml:"strip_encrypted_content"`
}

func Default() Config {
	return Config{
		Enabled:                     true,
		SourceFormats:               []string{"codex", "openai-response", "openai"},
		InterceptRuleMode:           RuleReasoningTokens,
		ReasoningMatchMode:          MatchFormula518NMinus2,
		ReasoningEquals:             []int{516, 1034, 1552},
		ContinuationMarkerText:      "Continue thinking...",
		InterceptStreaming:          true,
		InterceptNonStreaming:       true,
		NonStreamStatusCode:         http.StatusBadGateway,
		GuardRetryAttempts:          5,
		RetryUpstreamCapacityErrors: true,
		StreamAction:                StreamActionContinuation,
		LogMatch:                    true,
		StripEncryptedContent:       true,
	}
}

func Decode(configYAML []byte) (Config, error) {
	cfg := Default()
	if len(configYAML) == 0 {
		return cfg, nil
	}
	var raw rawConfig
	if err := yaml.Unmarshal(configYAML, &raw); err != nil {
		return Config{}, err
	}
	applyRaw(&cfg, raw)
	return normalize(cfg)
}

func applyRaw(cfg *Config, raw rawConfig) {
	if raw.Enabled != nil {
		cfg.Enabled = *raw.Enabled
	}
	cfg.SourceFormats = chooseStrings(raw.SourceFormats, cfg.SourceFormats)
	cfg.Models = normalizeStrings(raw.Models)
	cfg.ModelPrefixes = normalizeStrings(raw.ModelPrefixes)
	cfg.InterceptRuleMode = chooseString(raw.InterceptRuleMode, cfg.InterceptRuleMode)
	cfg.ReasoningMatchMode = chooseString(raw.ReasoningMatchMode, cfg.ReasoningMatchMode)
	cfg.ReasoningEquals = chooseInts(raw.ReasoningEquals, cfg.ReasoningEquals)
	cfg.ContinuationMarkerText = chooseString(raw.ContinuationMarkerText, cfg.ContinuationMarkerText)
	applyRawBooleans(cfg, raw)
}

func applyRawBooleans(cfg *Config, raw rawConfig) {
	if raw.InterceptStreaming != nil {
		cfg.InterceptStreaming = *raw.InterceptStreaming
	}
	if raw.InterceptNonStreaming != nil {
		cfg.InterceptNonStreaming = *raw.InterceptNonStreaming
	}
	if raw.NonStreamStatusCode != 0 {
		cfg.NonStreamStatusCode = raw.NonStreamStatusCode
	}
	if raw.GuardRetryAttempts != nil {
		cfg.GuardRetryAttempts = *raw.GuardRetryAttempts
	}
	if raw.RetryUpstreamCapacityErrors != nil {
		cfg.RetryUpstreamCapacityErrors = *raw.RetryUpstreamCapacityErrors
	}
	if raw.LogMatch != nil {
		cfg.LogMatch = *raw.LogMatch
	}
	if raw.StripEncryptedContent != nil {
		cfg.StripEncryptedContent = *raw.StripEncryptedContent
	}
	cfg.StreamAction = chooseString(raw.StreamAction, cfg.StreamAction)
}

func normalize(cfg Config) (Config, error) {
	cfg.InterceptRuleMode = strings.ToLower(strings.TrimSpace(cfg.InterceptRuleMode))
	cfg.ReasoningMatchMode = strings.ToLower(strings.TrimSpace(cfg.ReasoningMatchMode))
	cfg.StreamAction = strings.ToLower(strings.TrimSpace(cfg.StreamAction))
	cfg.SourceFormats = normalizeStrings(cfg.SourceFormats)
	if err := validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func validate(cfg Config) error {
	if cfg.GuardRetryAttempts < 0 {
		return fmt.Errorf("guard_retry_attempts must be >= 0")
	}
	if cfg.NonStreamStatusCode < 400 || cfg.NonStreamStatusCode > 599 {
		return fmt.Errorf("non_stream_status_code must be an HTTP error status")
	}
	if cfg.InterceptRuleMode != RuleReasoningTokens && cfg.InterceptRuleMode != RuleFinalOnlyHighXHigh {
		return fmt.Errorf("unsupported intercept_rule_mode: %s", cfg.InterceptRuleMode)
	}
	if cfg.ReasoningMatchMode != MatchManual && cfg.ReasoningMatchMode != MatchFormula518NMinus2 {
		return fmt.Errorf("unsupported reasoning_match_mode: %s", cfg.ReasoningMatchMode)
	}
	if cfg.StreamAction != StreamActionStrict502 && cfg.StreamAction != StreamActionContinuation {
		return fmt.Errorf("unsupported stream_action for plugin mode: %s", cfg.StreamAction)
	}
	return nil
}

func SupportsSourceFormat(cfg Config, format string) bool {
	format = strings.ToLower(strings.TrimSpace(format))
	for _, item := range cfg.SourceFormats {
		if strings.EqualFold(item, format) {
			return true
		}
	}
	return false
}

func SupportsModel(cfg Config, model string) bool {
	model = strings.TrimSpace(model)
	if len(cfg.Models) == 0 && len(cfg.ModelPrefixes) == 0 {
		return true
	}
	return matchesExact(cfg.Models, model) || matchesPrefix(cfg.ModelPrefixes, model)
}

func chooseString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func chooseStrings(values, fallback []string) []string {
	if len(values) == 0 {
		return append([]string(nil), fallback...)
	}
	return normalizeStrings(values)
}

func chooseInts(values, fallback []int) []int {
	if len(values) == 0 {
		return append([]int(nil), fallback...)
	}
	return append([]int(nil), values...)
}

type intList []int

func (values *intList) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode && node.Tag == "!!null" {
		*values = nil
		return nil
	}
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("reasoning_equals must be a list of integers")
	}
	out := make([]int, 0, len(node.Content))
	for index, item := range node.Content {
		parsed, err := parseIntNode(item)
		if err != nil {
			return fmt.Errorf("reasoning_equals[%d]: %w", index, err)
		}
		out = append(out, parsed)
	}
	*values = out
	return nil
}

func parseIntNode(node *yaml.Node) (int, error) {
	if node.Kind != yaml.ScalarNode {
		return 0, fmt.Errorf("must be an integer")
	}
	raw := strings.TrimSpace(node.Value)
	if raw == "" {
		return 0, fmt.Errorf("must not be empty")
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("must be an integer: %q", node.Value)
	}
	return value, nil
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func matchesExact(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), target) {
			return true
		}
	}
	return false
}

func matchesPrefix(values []string, target string) bool {
	for _, value := range values {
		if strings.HasPrefix(target, strings.TrimSpace(value)) {
			return true
		}
	}
	return false
}
