package audit

import (
	"testing"
	"time"
)

func makeStatsEntry(path, action string, ts time.Time) Entry {
	return Entry{Path: path, Action: action, Timestamp: ts}
}

func TestStats_Empty(t *testing.T) {
	result := Stats(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d entries", len(result))
	}
}

func TestStats_SinglePath(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeStatsEntry("secret/app", "added", now.Add(-2*time.Hour)),
		makeStatsEntry("secret/app", "updated", now.Add(-1*time.Hour)),
		makeStatsEntry("secret/app", "unchanged", now),
	}

	result := Stats(entries)
	if len(result) != 1 {
		t.Fatalf("expected 1 path stat, got %d", len(result))
	}

	s := result[0]
	if s.Path != "secret/app" {
		t.Errorf("unexpected path: %s", s.Path)
	}
	if s.TotalSyncs != 3 {
		t.Errorf("expected TotalSyncs=3, got %d", s.TotalSyncs)
	}
	if s.Added != 1 {
		t.Errorf("expected Added=1, got %d", s.Added)
	}
	if s.Updated != 1 {
		t.Errorf("expected Updated=1, got %d", s.Updated)
	}
	if s.Unchanged != 1 {
		t.Errorf("expected Unchanged=1, got %d", s.Unchanged)
	}
}

func TestStats_MultiplePaths(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeStatsEntry("secret/b", "added", now),
		makeStatsEntry("secret/a", "updated", now),
		makeStatsEntry("secret/a", "added", now.Add(-time.Hour)),
	}

	result := Stats(entries)
	if len(result) != 2 {
		t.Fatalf("expected 2 path stats, got %d", len(result))
	}
	// sorted by path
	if result[0].Path != "secret/a" {
		t.Errorf("expected first path to be secret/a, got %s", result[0].Path)
	}
	if result[0].TotalSyncs != 2 {
		t.Errorf("expected TotalSyncs=2 for secret/a, got %d", result[0].TotalSyncs)
	}
	if result[1].Path != "secret/b" {
		t.Errorf("expected second path to be secret/b, got %s", result[1].Path)
	}
}

func TestStats_FirstAndLastSeen(t *testing.T) {
	old := time.Now().Add(-24 * time.Hour)
	recent := time.Now()
	entries := []Entry{
		makeStatsEntry("secret/x", "added", recent),
		makeStatsEntry("secret/x", "updated", old),
	}

	result := Stats(entries)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if !result[0].FirstSeen.Equal(old) {
		t.Errorf("expected FirstSeen=%v, got %v", old, result[0].FirstSeen)
	}
	if !result[0].LastSeen.Equal(recent) {
		t.Errorf("expected LastSeen=%v, got %v", recent, result[0].LastSeen)
	}
}
