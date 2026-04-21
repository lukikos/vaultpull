package audit

import (
	"testing"
	"time"
)

var (
	now   = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	early = now.Add(-2 * time.Hour)
	late  = now.Add(2 * time.Hour)
)

func makeFilterEntry(path, action string, ts time.Time) LogEntry {
	return LogEntry{Timestamp: ts, Path: path, Action: action, Key: "KEY"}
}

func TestFilter_ByPath(t *testing.T) {
	entries := []LogEntry{
		makeFilterEntry("secret/a", "write", now),
		makeFilterEntry("secret/b", "write", now),
	}
	got := Filter(entries, FilterOptions{Path: "secret/a"})
	if len(got) != 1 || got[0].Path != "secret/a" {
		t.Fatalf("expected 1 entry for secret/a, got %d", len(got))
	}
}

func TestFilter_ByAction(t *testing.T) {
	entries := []LogEntry{
		makeFilterEntry("secret/a", "write", now),
		makeFilterEntry("secret/a", "skip", now),
	}
	got := Filter(entries, FilterOptions{Action: "skip"})
	if len(got) != 1 || got[0].Action != "skip" {
		t.Fatalf("expected 1 skip entry, got %d", len(got))
	}
}

func TestFilter_BySince(t *testing.T) {
	entries := []LogEntry{
		makeFilterEntry("secret/a", "write", early),
		makeFilterEntry("secret/a", "write", late),
	}
	got := Filter(entries, FilterOptions{Since: now})
	if len(got) != 1 || !got[0].Timestamp.Equal(late) {
		t.Fatalf("expected 1 entry at/after now, got %d", len(got))
	}
}

func TestFilter_ByUntil(t *testing.T) {
	entries := []LogEntry{
		makeFilterEntry("secret/a", "write", early),
		makeFilterEntry("secret/a", "write", late),
	}
	got := Filter(entries, FilterOptions{Until: now})
	if len(got) != 1 || !got[0].Timestamp.Equal(early) {
		t.Fatalf("expected 1 entry at/before now, got %d", len(got))
	}
}

func TestFilter_NoOpts_ReturnsAll(t *testing.T) {
	entries := []LogEntry{
		makeFilterEntry("secret/a", "write", early),
		makeFilterEntry("secret/b", "skip", late),
	}
	got := Filter(entries, FilterOptions{})
	if len(got) != 2 {
		t.Fatalf("expected all 2 entries, got %d", len(got))
	}
}

func TestFilter_Combined(t *testing.T) {
	entries := []LogEntry{
		makeFilterEntry("secret/a", "write", early),
		makeFilterEntry("secret/a", "write", late),
		makeFilterEntry("secret/b", "write", late),
	}
	got := Filter(entries, FilterOptions{Path: "secret/a", Since: now})
	if len(got) != 1 || got[0].Path != "secret/a" || !got[0].Timestamp.Equal(late) {
		t.Fatalf("expected 1 combined-filtered entry, got %d", len(got))
	}
}
