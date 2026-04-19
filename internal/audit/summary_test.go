package audit

import (
	"testing"
	"time"
)

func makeEntry(path, status string, t time.Time) Entry {
	return Entry{SecretPath: path, Status: status, Timestamp: t}
}

func TestSummarize_Empty(t *testing.T) {
	s := Summarize(nil)
	if s.Total != 0 {
		t.Errorf("expected 0 total, got %d", s.Total)
	}
	if s.String() != "No audit entries found." {
		t.Errorf("unexpected string: %s", s.String())
	}
}

func TestSummarize_Counts(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeEntry("secret/app", "success", now.Add(-2*time.Minute)),
		makeEntry("secret/app", "success", now.Add(-1*time.Minute)),
		makeEntry("secret/db", "error", now),
	}
	s := Summarize(entries)
	if s.Total != 3 {
		t.Errorf("expected 3, got %d", s.Total)
	}
	if s.Succeeded != 2 {
		t.Errorf("expected 2 succeeded, got %d", s.Succeeded)
	}
	if s.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", s.Failed)
	}
}

func TestSummarize_UniquePaths(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeEntry("secret/app", "success", now),
		makeEntry("secret/app", "success", now),
		makeEntry("secret/other", "success", now),
	}
	s := Summarize(entries)
	if len(s.Paths) != 2 {
		t.Errorf("expected 2 unique paths, got %d", len(s.Paths))
	}
}

func TestSummarize_LastSync(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	old := now.Add(-10 * time.Minute)
	entries := []Entry{
		makeEntry("secret/app", "success", old),
		makeEntry("secret/app", "success", now),
	}
	s := Summarize(entries)
	if !s.LastSync.Equal(now) {
		t.Errorf("expected last sync %v, got %v", now, s.LastSync)
	}
}
