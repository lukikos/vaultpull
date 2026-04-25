package audit_test

import (
	"testing"

	"github.com/yourusername/vaultpull/internal/audit"
)

func TestCompareSnapshots_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.CompareSnapshots(dir, "missing-a", "missing-b")
	if err == nil {
		t.Fatal("expected error for missing snapshots, got nil")
	}
}

func TestCompareSnapshots_DetectsAdded(t *testing.T) {
	dir := t.TempDir()

	if err := audit.SaveSnapshot(dir, "v1", map[string]string{"KEY_A": "alpha"}); err != nil {
		t.Fatalf("SaveSnapshot v1: %v", err)
	}
	if err := audit.SaveSnapshot(dir, "v2", map[string]string{"KEY_A": "alpha", "KEY_B": "beta"}); err != nil {
		t.Fatalf("SaveSnapshot v2: %v", err)
	}

	cmp, err := audit.CompareSnapshots(dir, "v1", "v2")
	if err != nil {
		t.Fatalf("CompareSnapshots: %v", err)
	}

	if cmp.From != "v1" || cmp.To != "v2" {
		t.Errorf("unexpected From/To: %q %q", cmp.From, cmp.To)
	}

	if len(cmp.Diffs) != 2 {
		t.Fatalf("expected 2 diff entries, got %d", len(cmp.Diffs))
	}
}

func TestCompareSnapshots_DetectsUpdated(t *testing.T) {
	dir := t.TempDir()

	if err := audit.SaveSnapshot(dir, "a", map[string]string{"FOO": "old"}); err != nil {
		t.Fatalf("SaveSnapshot a: %v", err)
	}
	if err := audit.SaveSnapshot(dir, "b", map[string]string{"FOO": "new"}); err != nil {
		t.Fatalf("SaveSnapshot b: %v", err)
	}

	cmp, err := audit.CompareSnapshots(dir, "a", "b")
	if err != nil {
		t.Fatalf("CompareSnapshots: %v", err)
	}

	if len(cmp.Diffs) != 1 || cmp.Diffs[0].Action != "updated" {
		t.Errorf("expected 1 updated diff, got %+v", cmp.Diffs)
	}
}

func TestCompareSnapshots_Summary(t *testing.T) {
	dir := t.TempDir()

	if err := audit.SaveSnapshot(dir, "s1", map[string]string{"A": "1", "B": "2"}); err != nil {
		t.Fatalf("SaveSnapshot s1: %v", err)
	}
	if err := audit.SaveSnapshot(dir, "s2", map[string]string{"A": "1", "C": "3"}); err != nil {
		t.Fatalf("SaveSnapshot s2: %v", err)
	}

	cmp, err := audit.CompareSnapshots(dir, "s1", "s2")
	if err != nil {
		t.Fatalf("CompareSnapshots: %v", err)
	}

	summary := cmp.Summary()
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}
