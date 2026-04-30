package audit

import (
	"testing"
	"time"
)

func makeLineageEntry(t time.Time, path, key, action, value string) Entry {
	return Entry{Timestamp: t, Path: path, Key: key, Action: action, Value: value}
}

func TestLineage_Empty(t *testing.T) {
	results := Lineage(nil, "secret/app", "")
	if len(results) != 0 {
		t.Fatalf("expected no results, got %d", len(results))
	}
}

func TestLineage_FiltersByPath(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeLineageEntry(now, "secret/app", "DB_PASS", "sync", "abc"),
		makeLineageEntry(now, "secret/other", "API_KEY", "sync", "xyz"),
	}
	results := Lineage(entries, "secret/app", "")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Key != "DB_PASS" {
		t.Errorf("unexpected key %q", results[0].Key)
	}
}

func TestLineage_FiltersByKey(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeLineageEntry(now, "secret/app", "DB_PASS", "sync", "abc"),
		makeLineageEntry(now, "secret/app", "API_KEY", "sync", "xyz"),
	}
	results := Lineage(entries, "secret/app", "API_KEY")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Key != "API_KEY" {
		t.Errorf("unexpected key %q", results[0].Key)
	}
}

func TestLineage_ChronologicalOrder(t *testing.T) {
	t1 := time.Now().Add(-2 * time.Hour)
	t2 := time.Now().Add(-1 * time.Hour)
	t3 := time.Now()
	entries := []Entry{
		makeLineageEntry(t3, "secret/app", "DB_PASS", "update", "v3"),
		makeLineageEntry(t1, "secret/app", "DB_PASS", "add", "v1"),
		makeLineageEntry(t2, "secret/app", "DB_PASS", "sync", "v2"),
	}
	results := Lineage(entries, "secret/app", "DB_PASS")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	h := results[0].History
	if len(h) != 3 {
		t.Fatalf("expected 3 history entries, got %d", len(h))
	}
	if h[0].Value != "v1" || h[1].Value != "v2" || h[2].Value != "v3" {
		t.Errorf("wrong order: %v %v %v", h[0].Value, h[1].Value, h[2].Value)
	}
}

func TestLineage_IgnoresNonRelevantActions(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeLineageEntry(now, "secret/app", "DB_PASS", "sync", "abc"),
		makeLineageEntry(now, "secret/app", "DB_PASS", "lint", ""),
	}
	results := Lineage(entries, "secret/app", "DB_PASS")
	if len(results[0].History) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(results[0].History))
	}
}

func TestLineage_MultipleKeysSortedByName(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeLineageEntry(now, "secret/app", "Z_KEY", "sync", "1"),
		makeLineageEntry(now, "secret/app", "A_KEY", "sync", "2"),
	}
	results := Lineage(entries, "secret/app", "")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Key != "A_KEY" || results[1].Key != "Z_KEY" {
		t.Errorf("wrong sort order: %q %q", results[0].Key, results[1].Key)
	}
}
