package metrics

import "testing"

func TestRecorderSnapshotAggregatesByModel(t *testing.T) {
	recorder := NewRecorder(2)
	recorder.Record(RequestRecord{Model: "gpt-5.5", Attempts: 2, RetryAttempts: 1, GuardMatches: 1, Intercepted: true})
	recorder.Record(RequestRecord{Model: "gpt-5.5"})
	recorder.Record(RequestRecord{Model: "gpt-5.4", Blocked: true, Intercepted: true, GuardMatches: 1})

	snapshot := recorder.Snapshot()
	if snapshot.TotalRequests != 3 || snapshot.InterceptedRequests != 2 || snapshot.BlockedRequests != 1 {
		t.Fatalf("snapshot counters = %#v", snapshot)
	}
	if snapshot.RetryAttempts != 1 || snapshot.GuardMatches != 2 {
		t.Fatalf("snapshot retry/match counters = %#v", snapshot)
	}
	if len(snapshot.Models) != 2 {
		t.Fatalf("models len = %d, want 2", len(snapshot.Models))
	}
	if snapshot.Models[0].Model != "gpt-5.5" || snapshot.Models[0].TotalRequests != 2 {
		t.Fatalf("first model = %#v, want gpt-5.5 with 2 requests", snapshot.Models[0])
	}
	if len(snapshot.Recent) != 2 {
		t.Fatalf("recent len = %d, want max 2", len(snapshot.Recent))
	}
	if snapshot.Recent[0].Model != "gpt-5.4" {
		t.Fatalf("recent[0] = %#v, want newest record first", snapshot.Recent[0])
	}
}
