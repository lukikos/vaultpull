package audit_test

import (
	"testing"

	"github.com/yourusername/vaultpull/internal/audit"
)

func makeBaseline(keys map[string]string) *audit.Baseline {
	return &audit.Baseline{Name: "test", Path: "secret/app", Keys: keys}
}

func TestBaselineDiff_AllUnchanged(t *testing.T) {
	b := makeBaseline(map[string]string{"A": "1", "B": "2"})
	result := audit.BaselineDiff(b, map[string]string{"A": "1", "B": "2"})
	for _, e := range result {
		if e.Action != "unchanged" {
			t.Errorf("expected unchanged for %q, got %q", e.Key, e.Action)
		}
	}
}

func TestBaselineDiff_DetectsAdded(t *testing.T) {
	b := makeBaseline(map[string]string{"A": "1"})
	result := audit.BaselineDiff(b, map[string]string{"A": "1", "B": "new"})
	found := false
	for _, e := range result {
		if e.Key == "B" && e.Action == "added" {
			found = true
		}
	}
	if !found {
		t.Error("expected B to be marked as added")
	}
}

func TestBaselineDiff_DetectsRemoved(t *testing.T) {
	b := makeBaseline(map[string]string{"A": "1", "B": "2"})
	result := audit.BaselineDiff(b, map[string]string{"A": "1"})
	for _, e := range result {
		if e.Key == "B" && e.Action != "removed" {
			t.Errorf("expected B removed, got %q", e.Action)
		}
	}
}

func TestBaselineDiff_DetectsChanged(t *testing.T) {
	b := makeBaseline(map[string]string{"TOKEN": "old"})
	result := audit.BaselineDiff(b, map[string]string{"TOKEN": "new"})
	if len(result) != 1 || result[0].Action != "changed" {
		t.Errorf("expected changed, got %+v", result)
	}
}

func TestBaselineDiff_SortedByKey(t *testing.T) {
	b := makeBaseline(map[string]string{"Z": "1", "A": "2", "M": "3"})
	result := audit.BaselineDiff(b, map[string]string{"Z": "1", "A": "2", "M": "3"})
	for i := 1; i < len(result); i++ {
		if result[i].Key < result[i-1].Key {
			t.Errorf("results not sorted at index %d", i)
		}
	}
}

func TestSummarizeBaselineDiff(t *testing.T) {
	b := makeBaseline(map[string]string{"A": "1", "B": "old"})
	result := audit.BaselineDiff(b, map[string]string{"B": "new", "C": "3"})
	summary := audit.SummarizeBaselineDiff(result)
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}
