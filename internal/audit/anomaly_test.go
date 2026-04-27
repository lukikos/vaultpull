package audit

import (
	"testing"
	"time"
)

func makeAnomalyEntry(path string, t time.Time) LogEntry {
	return LogEntry{Path: path, Action: "sync", Timestamp: t}
}

func TestDetectAnomalies_Empty(t *testing.T) {
	results := DetectAnomalies([]LogEntry{})
	if len(results) != 0 {
		t.Fatalf("expected no results, got %d", len(results))
	}
}

func TestDetectAnomalies_TooFewEntries(t *testing.T) {
	now := time.Now()
	entries := []LogEntry{
		makeAnomalyEntry("secret/app", now.Add(-2*time.Hour)),
		makeAnomalyEntry("secret/app", now.Add(-1*time.Hour)),
	}
	results := DetectAnomalies(entries)
	// fewer than 3 entries — should be skipped
	if len(results) != 0 {
		t.Fatalf("expected 0 results for path with <3 entries, got %d", len(results))
	}
}

func TestDetectAnomalies_NoAnomaly(t *testing.T) {
	now := time.Now()
	// Regular syncs every hour for 5 times
	entries := []LogEntry{
		makeAnomalyEntry("secret/app", now.Add(-4*time.Hour)),
		makeAnomalyEntry("secret/app", now.Add(-3*time.Hour)),
		makeAnomalyEntry("secret/app", now.Add(-2*time.Hour)),
		makeAnomalyEntry("secret/app", now.Add(-1*time.Hour)),
		makeAnomalyEntry("secret/app", now.Add(-1*time.Minute)),
	}
	results := DetectAnomalies(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].IsAnomaly {
		t.Errorf("expected no anomaly for regular sync pattern")
	}
}

func TestDetectAnomalies_DetectsOverdue(t *testing.T) {
	now := time.Now()
	// Syncs every minute historically, but last sync was 2 hours ago
	entries := []LogEntry{
		makeAnomalyEntry("secret/db", now.Add(-5*time.Minute)),
		makeAnomalyEntry("secret/db", now.Add(-4*time.Minute)),
		makeAnomalyEntry("secret/db", now.Add(-3*time.Minute)),
		makeAnomalyEntry("secret/db", now.Add(-2*time.Hour)),
	}
	results := DetectAnomalies(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].IsAnomaly {
		t.Errorf("expected anomaly for overdue path")
	}
	if results[0].Reason == "" {
		t.Errorf("expected non-empty reason")
	}
}

func TestDetectAnomalies_IgnoresNonSync(t *testing.T) {
	now := time.Now()
	entries := []LogEntry{
		{Path: "secret/app", Action: "read", Timestamp: now.Add(-3 * time.Hour)},
		{Path: "secret/app", Action: "delete", Timestamp: now.Add(-2 * time.Hour)},
		{Path: "secret/app", Action: "read", Timestamp: now.Add(-1 * time.Hour)},
	}
	results := DetectAnomalies(entries)
	if len(results) != 0 {
		t.Fatalf("expected 0 results when no sync actions, got %d", len(results))
	}
}

func TestDetectAnomalies_MultiplePathsIndependent(t *testing.T) {
	now := time.Now()
	entries := []LogEntry{
		makeAnomalyEntry("secret/a", now.Add(-3*time.Hour)),
		makeAnomalyEntry("secret/a", now.Add(-2*time.Hour)),
		makeAnomalyEntry("secret/a", now.Add(-1*time.Hour)),
		makeAnomalyEntry("secret/b", now.Add(-3*time.Minute)),
		makeAnomalyEntry("secret/b", now.Add(-2*time.Minute)),
		makeAnomalyEntry("secret/b", now.Add(-1*time.Minute)),
	}
	results := DetectAnomalies(entries)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
