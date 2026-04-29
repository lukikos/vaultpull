package audit_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func writeDigestEntries(t *testing.T, dir string, entries []audit.LogEntry) {
	t.Helper()
	logger, err := audit.NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for _, e := range entries {
		if err := logger.Record(e.Path, e.Action, e.Keys); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
}

func TestComputeDigests_Empty(t *testing.T) {
	dir := t.TempDir()
	report, err := audit.ComputeDigests(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(report.Entries))
	}
}

func TestComputeDigests_SinglePath(t *testing.T) {
	dir := t.TempDir()
	writeDigestEntries(t, dir, []audit.LogEntry{
		{Path: "secret/app", Action: "sync", Keys: []string{"DB_PASS", "API_KEY"}},
	})

	report, err := audit.ComputeDigests(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report.Entries))
	}
	e := report.Entries[0]
	if e.Path != "secret/app" {
		t.Errorf("path = %q, want %q", e.Path, "secret/app")
	}
	if e.Digest == "" {
		t.Error("digest should not be empty")
	}
	if e.ComputedAt.IsZero() {
		t.Error("computed_at should be set")
	}
}

func TestComputeDigests_DifferentPathsHaveDifferentDigests(t *testing.T) {
	dir := t.TempDir()
	writeDigestEntries(t, dir, []audit.LogEntry{
		{Path: "secret/app", Action: "sync", Keys: []string{"KEY_A"}},
		{Path: "secret/db", Action: "sync", Keys: []string{"KEY_B"}},
	})

	report, err := audit.ComputeDigests(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(report.Entries))
	}
	if report.Entries[0].Digest == report.Entries[1].Digest {
		t.Error("different paths should produce different digests")
	}
}

func TestSaveAndLoadDigests_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	writeDigestEntries(t, dir, []audit.LogEntry{
		{Path: "secret/app", Action: "sync", Keys: []string{"TOKEN"}},
	})

	report, err := audit.ComputeDigests(dir)
	if err != nil {
		t.Fatalf("ComputeDigests: %v", err)
	}
	if err := audit.SaveDigests(dir, report); err != nil {
		t.Fatalf("SaveDigests: %v", err)
	}

	loaded, err := audit.LoadDigests(dir)
	if err != nil {
		t.Fatalf("LoadDigests: %v", err)
	}
	if len(loaded.Entries) != len(report.Entries) {
		t.Fatalf("entry count mismatch: got %d, want %d", len(loaded.Entries), len(report.Entries))
	}
	if loaded.Entries[0].Digest != report.Entries[0].Digest {
		t.Errorf("digest mismatch after round-trip")
	}
}

func TestSaveDigests_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	report := audit.DigestReport{
		Entries: []audit.DigestEntry{
			{Path: "secret/x", Digest: "abc123", KeyCount: 1, ComputedAt: time.Now()},
		},
	}
	if err := audit.SaveDigests(dir, report); err != nil {
		t.Fatalf("SaveDigests: %v", err)
	}
	info, err := os.Stat(dir + "/.vaultpull_digests.json")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("file permissions = %o, want 0600", info.Mode().Perm())
	}
}

func TestLoadDigests_NoFile(t *testing.T) {
	dir := t.TempDir()
	report, err := audit.LoadDigests(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Entries) != 0 {
		t.Errorf("expected empty report, got %d entries", len(report.Entries))
	}
}
