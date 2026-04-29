package audit

import (
	"testing"
	"time"
)

func makeThrottleEntry(path, action string, ts time.Time) Entry {
	return Entry{Path: path, Action: action, Timestamp: ts}
}

func TestCheckThrottle_Empty(t *testing.T) {
	results := CheckThrottle(nil, DefaultThrottleConfig(), time.Now())
	if len(results) != 0 {
		t.Fatalf("expected no results, got %d", len(results))
	}
}

func TestCheckThrottle_BelowLimit(t *testing.T) {
	now := time.Now()
	cfg := ThrottleConfig{MaxSyncs: 5, Window: time.Hour}
	var entries []Entry
	for i := 0; i < 3; i++ {
		entries = append(entries, makeThrottleEntry("secret/app", "sync", now.Add(-time.Duration(i)*time.Minute)))
	}
	results := CheckThrottle(entries, cfg, now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Exceeds {
		t.Error("expected Exceeds=false")
	}
	if results[0].SyncCount != 3 {
		t.Errorf("expected SyncCount=3, got %d", results[0].SyncCount)
	}
}

func TestCheckThrottle_ExceedsLimit(t *testing.T) {
	now := time.Now()
	cfg := ThrottleConfig{MaxSyncs: 3, Window: time.Hour}
	var entries []Entry
	for i := 0; i < 5; i++ {
		entries = append(entries, makeThrottleEntry("secret/db", "sync", now.Add(-time.Duration(i)*time.Minute)))
	}
	results := CheckThrottle(entries, cfg, now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Exceeds {
		t.Error("expected Exceeds=true")
	}
	if results[0].Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestCheckThrottle_IgnoresOldEntries(t *testing.T) {
	now := time.Now()
	cfg := ThrottleConfig{MaxSyncs: 2, Window: time.Hour}
	entries := []Entry{
		makeThrottleEntry("secret/app", "sync", now.Add(-90*time.Minute)), // outside window
		makeThrottleEntry("secret/app", "sync", now.Add(-30*time.Minute)),
		makeThrottleEntry("secret/app", "sync", now.Add(-10*time.Minute)),
	}
	results := CheckThrottle(entries, cfg, now)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Only 2 entries fall within the window — exactly at limit, not exceeding.
	if results[0].Exceeds {
		t.Errorf("expected Exceeds=false, got true (count=%d)", results[0].SyncCount)
	}
}

func TestCheckThrottle_IgnoresNonSyncActions(t *testing.T) {
	now := time.Now()
	cfg := ThrottleConfig{MaxSyncs: 1, Window: time.Hour}
	entries := []Entry{
		makeThrottleEntry("secret/app", "read", now.Add(-5*time.Minute)),
		makeThrottleEntry("secret/app", "delete", now.Add(-3*time.Minute)),
	}
	results := CheckThrottle(entries, cfg, now)
	if len(results) != 0 {
		t.Fatalf("expected 0 results (no sync actions), got %d", len(results))
	}
}

func TestCheckThrottle_MultiplePathsIndependent(t *testing.T) {
	now := time.Now()
	cfg := ThrottleConfig{MaxSyncs: 2, Window: time.Hour}
	entries := []Entry{
		makeThrottleEntry("secret/a", "sync", now.Add(-10*time.Minute)),
		makeThrottleEntry("secret/a", "sync", now.Add(-20*time.Minute)),
		makeThrottleEntry("secret/a", "sync", now.Add(-30*time.Minute)), // exceeds
		makeThrottleEntry("secret/b", "sync", now.Add(-5*time.Minute)),  // fine
	}
	results := CheckThrottle(entries, cfg, now)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Path == "secret/a" && !r.Exceeds {
			t.Error("expected secret/a to exceed limit")
		}
		if r.Path == "secret/b" && r.Exceeds {
			t.Error("expected secret/b to be within limit")
		}
	}
}
