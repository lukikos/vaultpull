package audit

import (
	"testing"
	"time"
)

func makeTrendEntry(path, action string, ts time.Time) Entry {
	return Entry{Path: path, Action: action, Timestamp: ts, Key: "K"}
}

func TestTrend_Empty(t *testing.T) {
	results := Trend(nil, time.Now().Add(-24*time.Hour))
	if len(results) != 0 {
		t.Fatalf("expected no results, got %d", len(results))
	}
}

func TestTrend_IgnoresNonSyncActions(t *testing.T) {
	now := time.Now().UTC()
	entries := []Entry{
		makeTrendEntry("secret/app", "read", now),
		makeTrendEntry("secret/app", "delete", now),
	}
	results := Trend(entries, now.Add(-time.Hour))
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestTrend_SinglePath(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	entries := []Entry{
		makeTrendEntry("secret/app", "sync", base),
		makeTrendEntry("secret/app", "sync", base),
		makeTrendEntry("secret/app", "sync", base.Add(24*time.Hour)),
	}
	results := Trend(entries, base.Add(-time.Hour))
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Path != "secret/app" {
		t.Errorf("unexpected path: %s", r.Path)
	}
	if len(r.Points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(r.Points))
	}
	if r.PeakCount != 2 {
		t.Errorf("expected peak count 2, got %d", r.PeakCount)
	}
	if r.AvgPerDay != 1.5 {
		t.Errorf("expected avg 1.5, got %f", r.AvgPerDay)
	}
}

func TestTrend_MultiplePathsIndependent(t *testing.T) {
	base := time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC)
	entries := []Entry{
		makeTrendEntry("secret/a", "sync", base),
		makeTrendEntry("secret/b", "sync", base),
		makeTrendEntry("secret/b", "sync", base),
	}
	results := Trend(entries, base.Add(-time.Hour))
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Path == "secret/b" && r.PeakCount != 2 {
			t.Errorf("expected peak 2 for secret/b, got %d", r.PeakCount)
		}
	}
}

func TestTrend_SinceFiltersOldEntries(t *testing.T) {
	old := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	recent := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	entries := []Entry{
		makeTrendEntry("secret/app", "sync", old),
		makeTrendEntry("secret/app", "sync", recent),
	}
	results := Trend(entries, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if len(results[0].Points) != 1 {
		t.Errorf("expected 1 point after filter, got %d", len(results[0].Points))
	}
}

func TestTrend_SlopePositive(t *testing.T) {
	base := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	var entries []Entry
	for i := 0; i < 3; i++ {
		for j := 0; j <= i; j++ {
			entries = append(entries, makeTrendEntry("secret/grow", "sync", base.Add(time.Duration(i)*24*time.Hour)))
		}
	}
	results := Trend(entries, base.Add(-time.Hour))
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].TrendSlope <= 0 {
		t.Errorf("expected positive slope, got %f", results[0].TrendSlope)
	}
}
