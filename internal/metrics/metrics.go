package metrics

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const unknownModel = "(unknown)"

type RequestRecord struct {
	Timestamp      time.Time `json:"timestamp"`
	Model          string    `json:"model"`
	Format         string    `json:"format,omitempty"`
	Stream         bool      `json:"stream"`
	Intercepted    bool      `json:"intercepted"`
	Blocked        bool      `json:"blocked"`
	Attempts       int       `json:"attempts"`
	RetryAttempts  int       `json:"retry_attempts"`
	GuardMatches   int       `json:"guard_matches"`
	Mode           string    `json:"mode,omitempty"`
	Reason         string    `json:"reason,omitempty"`
	ReasoningToken *int      `json:"reasoning_tokens,omitempty"`
	ErrorCode      string    `json:"error_code,omitempty"`
	HTTPStatus     int       `json:"http_status,omitempty"`
}

type Snapshot struct {
	StartedAt           time.Time       `json:"started_at"`
	GeneratedAt         time.Time       `json:"generated_at"`
	TotalRequests       int64           `json:"total_requests"`
	InterceptedRequests int64           `json:"intercepted_requests"`
	BlockedRequests     int64           `json:"blocked_requests"`
	RetryAttempts       int64           `json:"retry_attempts"`
	GuardMatches        int64           `json:"guard_matches"`
	InterceptRatio      float64         `json:"intercept_ratio"`
	Models              []ModelStats    `json:"models"`
	Recent              []RequestRecord `json:"recent"`
}

type ModelStats struct {
	Model               string    `json:"model"`
	TotalRequests       int64     `json:"total_requests"`
	InterceptedRequests int64     `json:"intercepted_requests"`
	BlockedRequests     int64     `json:"blocked_requests"`
	RetryAttempts       int64     `json:"retry_attempts"`
	GuardMatches        int64     `json:"guard_matches"`
	InterceptRatio      float64   `json:"intercept_ratio"`
	LastSeenAt          time.Time `json:"last_seen_at"`
}

type Recorder struct {
	mu                  sync.Mutex
	startedAt           time.Time
	totalRequests       int64
	interceptedRequests int64
	blockedRequests     int64
	retryAttempts       int64
	guardMatches        int64
	models              map[string]*ModelStats
	recent              []RequestRecord
	maxRecent           int
}

func NewRecorder(maxRecent int) *Recorder {
	if maxRecent <= 0 {
		maxRecent = 100
	}
	return &Recorder{
		startedAt: time.Now(),
		models:    map[string]*ModelStats{},
		maxRecent: maxRecent,
	}
}

func (r *Recorder) Record(record RequestRecord) {
	if r == nil {
		return
	}
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}
	record.Model = normalizeModel(record.Model)
	if record.Attempts <= 0 {
		record.Attempts = 1
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.totalRequests++
	if record.Intercepted {
		r.interceptedRequests++
	}
	if record.Blocked {
		r.blockedRequests++
	}
	r.retryAttempts += int64(maxInt(record.RetryAttempts, 0))
	r.guardMatches += int64(maxInt(record.GuardMatches, 0))

	model := r.models[record.Model]
	if model == nil {
		model = &ModelStats{Model: record.Model}
		r.models[record.Model] = model
	}
	model.TotalRequests++
	if record.Intercepted {
		model.InterceptedRequests++
	}
	if record.Blocked {
		model.BlockedRequests++
	}
	model.RetryAttempts += int64(maxInt(record.RetryAttempts, 0))
	model.GuardMatches += int64(maxInt(record.GuardMatches, 0))
	model.LastSeenAt = record.Timestamp
	model.InterceptRatio = ratio(model.InterceptedRequests, model.TotalRequests)

	r.recent = append(r.recent, record)
	if overflow := len(r.recent) - r.maxRecent; overflow > 0 {
		copy(r.recent, r.recent[overflow:])
		r.recent = r.recent[:r.maxRecent]
	}
}

func (r *Recorder) Snapshot() Snapshot {
	if r == nil {
		now := time.Now()
		return Snapshot{StartedAt: now, GeneratedAt: now}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	models := make([]ModelStats, 0, len(r.models))
	for _, model := range r.models {
		models = append(models, *model)
	}
	sort.Slice(models, func(i, j int) bool {
		if models[i].TotalRequests != models[j].TotalRequests {
			return models[i].TotalRequests > models[j].TotalRequests
		}
		return models[i].Model < models[j].Model
	})

	recent := append([]RequestRecord(nil), r.recent...)
	for i, j := 0, len(recent)-1; i < j; i, j = i+1, j-1 {
		recent[i], recent[j] = recent[j], recent[i]
	}

	return Snapshot{
		StartedAt:           r.startedAt,
		GeneratedAt:         time.Now(),
		TotalRequests:       r.totalRequests,
		InterceptedRequests: r.interceptedRequests,
		BlockedRequests:     r.blockedRequests,
		RetryAttempts:       r.retryAttempts,
		GuardMatches:        r.guardMatches,
		InterceptRatio:      ratio(r.interceptedRequests, r.totalRequests),
		Models:              models,
		Recent:              recent,
	}
}

func normalizeModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return unknownModel
	}
	return model
}

func ratio(value, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(value) / float64(total)
}

func maxInt(value, floor int) int {
	if value < floor {
		return floor
	}
	return value
}
