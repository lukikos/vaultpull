package audit

import (
	"fmt"
	"testing"
	"time"
)

func makeVelocityEntry(path, action string, daysAgo int) Entry {
	t := time.Now().UTC().AddDate(0, 0, -daysAgo).Format(time.RFC3339)
	return Entry{Timestamp: t, Path: path, Action: action, Keys: []string{"K"}}
}

func TestVelocity_Empty(t *testing.T) {
	results := Velocity(nil, 7)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestVelocity_ZeroWindow(t *testing.T) {
	entries := []Entry{makeVelocityEntry("secret/app", "sync", 1)}
	results := Velocity(entries, 0)
	if len(results) != 0 {
		t.Fatalf("expected 0 results for zero window, got %d", len(results))
	}
}

func TestVelocity_IgnoresNonSyncActions(t *testing.T) {
	entries := []Entry{
		makeVelocityEntry("secret/app", "read", 1),
		makeVelocityEntry("secret/app", "delete", 2),
	}
	results := Velocity(entries, 7)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestVelocity_SinglePath(t *testing.T) {
	entries := []Entry{
		makeVelocityEntry("secret/app", "sync", 1),
		makeVelocityEntry("secret/app", "sync", 2),
		makeVelocityEntry("secret/app", "sync", 3),
	}
	results := Velocity(entries, 7)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "secret/app" {
		t.Errorf("unexpected path: %s", results[0].Path)
	}
	if results[0].SyncCount != 3 {
		t.Errorf("expected SyncCount 3, got %d", results[0].SyncCount)
	}
}

func TestVelocity_ExcludesOldEntries(t *testing.T) {
	entries := []Entry{
		makeVelocityEntry("secret/app", "sync", 1),
		makeVelocityEntry("secret/app", "sync", 30), // outside 7-day window
	}
	results := Velocity(entries, 7)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].SyncCount != 1 {
		t.Errorf("expected SyncCount 1, got %d", results[0].SyncCount)
	}
}

func TestVelocity_MultiplePathsSortedByRate(t *testing.T) {
	var entries []Entry
	for i := 0; i < 5; i++ {
		entries = append(entries, makeVelocityEntry("secret/high", "sync", i))
	}
	entries = append(entries, makeVelocityEntry("secret/low", "sync", 1))

	results := Velocity(entries, 7)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Path != "secret/high" {
		t.Errorf("expected secret/high first, got %s", results[0].Path)
	}
}

func TestVelocity_PeakDayDetected(t *testing.T) {
	// Two syncs on same day (today), one on yesterday
	today := time.Now().UTC().Format("2006-01-02")
	e1 := Entry{Timestamp: fmt.Sprintf("%sT10:00:00Z", today), Path: "secret/app", Action: "sync", Keys: []string{"K"}}
	e2 := Entry{Timestamp: fmt.Sprintf("%sT14:00:00Z", today), Path: "secret/app", Action: "sync", Keys: []string{"K"}}
	e3 := makeVelocityEntry("secret/app", "sync", 1)

	results := Velocity([]Entry{e1, e2, e3}, 7)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].PeakDay != today {
		t.Errorf("expected peak day %s, got %s", today, results[0].PeakDay)
	}
	if results[0].PeakDayCount != 2 {
		t.Errorf("expected peak count 2, got %d", results[0].PeakDayCount)
	}
}
