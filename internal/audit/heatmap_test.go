package audit

import (
	"testing"
)

func TestHeatmap_Empty(t *testing.T) {
	result := Heatmap(nil)
	if result != nil {
		t.Fatalf("expected nil result for empty entries, got %v", result)
	}
}

func TestHeatmap_IgnoresNonSyncActions(t *testing.T) {
	entries := []Entry{
		{Path: "secret/app", Action: "read", Timestamp: makeHeatmapEntry("secret/app", 10).Timestamp},
		{Path: "secret/app", Action: "delete", Timestamp: makeHeatmapEntry("secret/app", 14).Timestamp},
	}
	result := Heatmap(entries)
	if result != nil {
		t.Fatalf("expected nil when no sync actions, got %v", result)
	}
}

func TestHeatmap_SinglePath(t *testing.T) {
	entries := []Entry{
		makeHeatmapEntry("secret/app", 9),
		makeHeatmapEntry("secret/app", 9),
		makeHeatmapEntry("secret/app", 14),
	}
	results := Heatmap(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Path != "secret/app" {
		t.Errorf("expected path secret/app, got %s", r.Path)
	}
	if len(r.Buckets) != 24 {
		t.Errorf("expected 24 buckets, got %d", len(r.Buckets))
	}
	if r.Buckets[9].Count != 2 {
		t.Errorf("expected hour 9 count=2, got %d", r.Buckets[9].Count)
	}
	if r.Buckets[14].Count != 1 {
		t.Errorf("expected hour 14 count=1, got %d", r.Buckets[14].Count)
	}
	if r.PeakHour != 9 {
		t.Errorf("expected peak hour 9, got %d", r.PeakHour)
	}
	if r.PeakCount != 2 {
		t.Errorf("expected peak count 2, got %d", r.PeakCount)
	}
}

func TestHeatmap_MultiplePathsSorted(t *testing.T) {
	entries := []Entry{
		makeHeatmapEntry("secret/zz", 3),
		makeHeatmapEntry("secret/aa", 22),
		makeHeatmapEntry("secret/mm", 11),
	}
	results := Heatmap(entries)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Path != "secret/aa" || results[1].Path != "secret/mm" || results[2].Path != "secret/zz" {
		t.Errorf("results not sorted by path: %v", []string{results[0].Path, results[1].Path, results[2].Path})
	}
}

func TestHeatmap_ZeroCountBuckets(t *testing.T) {
	entries := []Entry{
		makeHeatmapEntry("secret/app", 0),
	}
	results := Heatmap(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	for h := 1; h < 24; h++ {
		if results[0].Buckets[h].Count != 0 {
			t.Errorf("expected hour %d count=0, got %d", h, results[0].Buckets[h].Count)
		}
	}
}
