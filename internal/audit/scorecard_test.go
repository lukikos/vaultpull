package audit

import (
	"testing"
	"time"
)

func makeScorecardEntry(path, action string, daysAgo int) Entry {
	return Entry{
		Timestamp: time.Now().AddDate(0, 0, -daysAgo),
		Path:      path,
		Action:    action,
		Key:       "KEY",
	}
}

func TestScorecard_Empty(t *testing.T) {
	results := Scorecard(nil)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestScorecard_HealthyPath(t *testing.T) {
	entries := []Entry{
		makeScorecardEntry("secret/app", "write", 1),
		makeScorecardEntry("secret/app", "write", 3),
	}
	results := Scorecard(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Score != 100 {
		t.Errorf("expected score 100, got %d (notes: %v)", r.Score, r.Notes)
	}
	if r.RecentSyncs != 2 {
		t.Errorf("expected 2 recent syncs, got %d", r.RecentSyncs)
	}
}

func TestScorecard_StalePath(t *testing.T) {
	entries := []Entry{
		makeScorecardEntry("secret/old", "write", 45),
	}
	results := Scorecard(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	r := results[0]
	if r.Score >= 80 {
		t.Errorf("expected low score for stale path, got %d", r.Score)
	}
	if len(r.Notes) == 0 {
		t.Error("expected notes for stale path")
	}
}

func TestScorecard_PathWithErrors(t *testing.T) {
	entries := []Entry{
		makeScorecardEntry("secret/broken", "write", 1),
		makeScorecardEntry("secret/broken", "error", 1),
	}
	results := Scorecard(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	r := results[0]
	if !r.HasErrors {
		t.Error("expected HasErrors to be true")
	}
	if r.Score >= 100 {
		t.Errorf("expected score < 100 for path with errors, got %d", r.Score)
	}
}

func TestScorecard_SortedByScore(t *testing.T) {
	entries := []Entry{
		makeScorecardEntry("secret/healthy", "write", 1),
		makeScorecardEntry("secret/stale", "write", 60),
	}
	results := Scorecard(entries)
	if len(results) != 2 {
		t.Fatalf("expected 2 results")
	}
	if results[0].Score > results[1].Score {
		t.Errorf("expected results sorted ascending by score, got %d then %d",
			results[0].Score, results[1].Score)
	}
}

func TestScorecard_MultiplePathsIndependent(t *testing.T) {
	entries := []Entry{
		makeScorecardEntry("secret/a", "write", 1),
		makeScorecardEntry("secret/b", "write", 1),
		makeScorecardEntry("secret/b", "write", 2),
	}
	results := Scorecard(entries)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.TotalSyncs == 0 {
			t.Errorf("path %s has 0 total syncs", r.Path)
		}
	}
}
