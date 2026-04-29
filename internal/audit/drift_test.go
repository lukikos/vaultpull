package audit

import (
	"testing"
	"time"
)

func makeDriftEntry(path, action string, ts time.Time) Entry {
	return Entry{Path: path, Action: action, Timestamp: ts}
}

func TestDetectDrift_Empty(t *testing.T) {
	results := DetectDrift(nil, DefaultDriftConfig(), time.Now())
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestDetectDrift_FreshPath(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeDriftEntry("secret/app", "sync", now.Add(-1*time.Hour)),
	}
	cfg := DefaultDriftConfig()
	results := DetectDrift(entries, cfg, now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].HasDrift {
		t.Errorf("expected no drift for fresh path, got reason: %s", results[0].Reason)
	}
}

func TestDetectDrift_StalePath(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeDriftEntry("secret/old", "sync", now.Add(-72*time.Hour)),
	}
	cfg := DefaultDriftConfig() // MaxAgeHours = 48
	results := DetectDrift(entries, cfg, now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].HasDrift {
		t.Error("expected drift for stale path")
	}
	if results[0].Reason == "" {
		t.Error("expected non-empty reason")
	}
}

func TestDetectDrift_IgnoresNonSyncActions(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeDriftEntry("secret/app", "read", now.Add(-1*time.Hour)),
		makeDriftEntry("secret/app", "export", now.Add(-2*time.Hour)),
	}
	results := DetectDrift(entries, DefaultDriftConfig(), now)
	if len(results) != 0 {
		t.Errorf("expected 0 results (no sync entries), got %d", len(results))
	}
}

func TestDetectDrift_MultiplePathsSorted(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeDriftEntry("secret/z", "sync", now.Add(-1*time.Hour)),
		makeDriftEntry("secret/a", "sync", now.Add(-1*time.Hour)),
		makeDriftEntry("secret/m", "sync", now.Add(-1*time.Hour)),
	}
	results := DetectDrift(entries, DefaultDriftConfig(), now)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Path != "secret/a" || results[1].Path != "secret/m" || results[2].Path != "secret/z" {
		t.Errorf("results not sorted: %v", results)
	}
}

func TestDetectDrift_UsesLatestSyncTime(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeDriftEntry("secret/app", "sync", now.Add(-100*time.Hour)),
		makeDriftEntry("secret/app", "sync", now.Add(-1*time.Hour)),
	}
	cfg := DefaultDriftConfig()
	results := DetectDrift(entries, cfg, now)
	if results[0].HasDrift {
		t.Error("expected no drift: latest sync is recent")
	}
	if results[0].SyncCount != 2 {
		t.Errorf("expected sync count 2, got %d", results[0].SyncCount)
	}
}
