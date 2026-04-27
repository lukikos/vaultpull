package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeRetentionEntries(t *testing.T, logPath string, entries []Entry) {
	t.Helper()
	l, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for _, e := range entries {
		if err := l.Record(e.Path, e.Action, e.Keys); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
}

func TestSaveRetentionPolicy_NegativeValues(t *testing.T) {
	dir := t.TempDir()
	err := SaveRetentionPolicy(dir, RetentionPolicy{MaxAgeDays: -1})
	if err == nil {
		t.Fatal("expected error for negative MaxAgeDays")
	}
}

func TestSaveRetentionPolicy_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	policy := RetentionPolicy{MaxAgeDays: 30, MaxEntries: 500}
	if err := SaveRetentionPolicy(dir, policy); err != nil {
		t.Fatalf("SaveRetentionPolicy: %v", err)
	}
	loaded, err := LoadRetentionPolicy(dir)
	if err != nil {
		t.Fatalf("LoadRetentionPolicy: %v", err)
	}
	if loaded.MaxAgeDays != 30 || loaded.MaxEntries != 500 {
		t.Errorf("got %+v, want {30 500}", loaded)
	}
}

func TestLoadRetentionPolicy_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadRetentionPolicy(dir)
	if err == nil {
		t.Fatal("expected error when no policy file exists")
	}
}

func TestEnforceRetention_NoFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.log")
	result, err := EnforceRetention(logPath, RetentionPolicy{MaxAgeDays: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Removed != 0 {
		t.Errorf("expected 0 removed, got %d", result.Removed)
	}
}

func TestEnforceRetention_RemovesOldEntries(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.log")
	now := time.Now()
	old := Entry{Path: "secret/old", Action: "sync", Keys: []string{"K"}, Timestamp: now.AddDate(0, 0, -10)}
	recent := Entry{Path: "secret/new", Action: "sync", Keys: []string{"K"}, Timestamp: now}

	f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0o600)
	writeJSONEntry(t, f, old)
	writeJSONEntry(t, f, recent)
	f.Close()

	result, err := EnforceRetention(logPath, RetentionPolicy{MaxAgeDays: 7})
	if err != nil {
		t.Fatalf("EnforceRetention: %v", err)
	}
	if result.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", result.Removed)
	}
	if result.Retained != 1 {
		t.Errorf("expected 1 retained, got %d", result.Retained)
	}
}

func TestEnforceRetention_MaxEntries(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "audit.log")
	f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0o600)
	now := time.Now()
	for i := 0; i < 5; i++ {
		e := Entry{Path: "secret/app", Action: "sync", Keys: []string{"K"}, Timestamp: now.Add(time.Duration(i) * time.Minute)}
		writeJSONEntry(t, f, e)
	}
	f.Close()

	result, err := EnforceRetention(logPath, RetentionPolicy{MaxEntries: 3})
	if err != nil {
		t.Fatalf("EnforceRetention: %v", err)
	}
	if result.Retained != 3 {
		t.Errorf("expected 3 retained, got %d", result.Retained)
	}
	if result.Removed != 2 {
		t.Errorf("expected 2 removed, got %d", result.Removed)
	}
}

func TestSaveRetentionPolicy_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	if err := SaveRetentionPolicy(dir, RetentionPolicy{MaxAgeDays: 7}); err != nil {
		t.Fatalf("SaveRetentionPolicy: %v", err)
	}
	info, err := os.Stat(retentionFilePath(dir))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected 0600, got %o", info.Mode().Perm())
	}
}
