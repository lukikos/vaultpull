package audit

import (
	"testing"
	"time"
)

func makeCouplingEntry(t time.Time, path, action string) Entry {
	return Entry{Timestamp: t, Path: path, Action: action, Key: "k"}
}

func TestDetectCoupling_Empty(t *testing.T) {
	results := DetectCoupling(nil, 0.0)
	if len(results) != 0 {
		t.Fatalf("expected empty, got %d", len(results))
	}
}

func TestDetectCoupling_NoSyncActions(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeCouplingEntry(now, "secret/a", "read"),
		makeCouplingEntry(now, "secret/b", "read"),
	}
	results := DetectCoupling(entries, 0.0)
	if len(results) != 0 {
		t.Fatalf("expected empty for non-sync actions, got %d", len(results))
	}
}

func TestDetectCoupling_DetectsPair(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	entries := []Entry{
		makeCouplingEntry(base, "secret/a", "sync"),
		makeCouplingEntry(base, "secret/b", "sync"),
		makeCouplingEntry(base.Add(2*time.Second), "secret/a", "sync"),
		makeCouplingEntry(base.Add(2*time.Second), "secret/b", "sync"),
	}
	results := DetectCoupling(entries, 0.0)
	if len(results) != 1 {
		t.Fatalf("expected 1 pair, got %d", len(results))
	}
	if results[0].PathA != "secret/a" || results[0].PathB != "secret/b" {
		t.Errorf("unexpected pair: %+v", results[0])
	}
	if results[0].CoOccurs != 2 {
		t.Errorf("expected co-occurs=2, got %d", results[0].CoOccurs)
	}
}

func TestDetectCoupling_MinSupportFilters(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	entries := []Entry{
		makeCouplingEntry(base, "secret/a", "sync"),
		makeCouplingEntry(base, "secret/b", "sync"),
		makeCouplingEntry(base.Add(2*time.Second), "secret/a", "sync"),
	}
	// pair appears in 1 of 2 batches → support=0.5
	results := DetectCoupling(entries, 0.8)
	if len(results) != 0 {
		t.Fatalf("expected 0 results above minSupport=0.8, got %d", len(results))
	}
	results = DetectCoupling(entries, 0.4)
	if len(results) != 1 {
		t.Fatalf("expected 1 result at minSupport=0.4, got %d", len(results))
	}
}

func TestDetectCoupling_SortedByCoOccurs(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	var entries []Entry
	for i := 0; i < 3; i++ {
		t2 := base.Add(time.Duration(i) * 2 * time.Second)
		entries = append(entries, makeCouplingEntry(t2, "secret/a", "sync"))
		entries = append(entries, makeCouplingEntry(t2, "secret/b", "sync"))
		if i < 1 {
			entries = append(entries, makeCouplingEntry(t2, "secret/c", "sync"))
		}
	}
	results := DetectCoupling(entries, 0.0)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].CoOccurs < results[len(results)-1].CoOccurs {
		t.Error("results not sorted descending by co-occurs")
	}
}
