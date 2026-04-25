package audit_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func writeArchiveEntries(t *testing.T, logPath string, entries []audit.Entry) {
	t.Helper()
	logger, err := audit.NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for _, e := range entries {
		if err := logger.Record(e.Path, e.Key, e.Action, e.Value); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
}

func TestArchive_EmptyName(t *testing.T) {
	dir := t.TempDir()
	logPath := dir + "/audit.log"
	err := audit.Archive(logPath, dir, "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestArchive_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	logPath := dir + "/audit.log"

	writeArchiveEntries(t, logPath, []audit.Entry{
		{Path: "secret/app", Key: "DB_PASS", Action: "added", Value: "secret", Timestamp: time.Now().UTC()},
	})

	if err := audit.Archive(logPath, dir, "before-deploy"); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	archivePath := dir + "/.vaultpull-archive-before-deploy.json"
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("expected archive file to exist: %v", err)
	}
}

func TestArchive_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	logPath := dir + "/audit.log"
	writeArchiveEntries(t, logPath, []audit.Entry{
		{Path: "secret/app", Key: "API_KEY", Action: "added", Value: "xyz", Timestamp: time.Now().UTC()},
	})

	if err := audit.Archive(logPath, dir, "perm-test"); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	info, err := os.Stat(dir + "/.vaultpull-archive-perm-test.json")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

func TestLoadArchive_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.LoadArchive(dir, "ghost")
	if err == nil {
		t.Fatal("expected error for missing archive")
	}
}

func TestLoadArchive_PreservesEntries(t *testing.T) {
	dir := t.TempDir()
	logPath := dir + "/audit.log"

	writeArchiveEntries(t, logPath, []audit.Entry{
		{Path: "secret/app", Key: "DB_PASS", Action: "added", Value: "s3cr3t", Timestamp: time.Now().UTC()},
		{Path: "secret/app", Key: "API_KEY", Action: "updated", Value: "newkey", Timestamp: time.Now().UTC()},
	})

	if err := audit.Archive(logPath, dir, "v1"); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	loaded, err := audit.LoadArchive(dir, "v1")
	if err != nil {
		t.Fatalf("LoadArchive: %v", err)
	}

	if loaded.Name != "v1" {
		t.Errorf("expected name v1, got %q", loaded.Name)
	}
	if len(loaded.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Key != "DB_PASS" {
		t.Errorf("expected first key DB_PASS, got %q", loaded.Entries[0].Key)
	}
}
