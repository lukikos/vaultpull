package audit_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func TestSaveSchema_InvalidVersion(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveSchema(dir, audit.SchemaVersion{Version: 0, Fields: []string{"key"}})
	if err == nil {
		t.Fatal("expected error for version=0")
	}
}

func TestSaveSchema_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	sv := audit.SchemaVersion{Version: 1, Fields: []string{"key", "path", "action"}}
	if err := audit.SaveSchema(dir, sv); err != nil {
		t.Fatalf("SaveSchema: %v", err)
	}
	info, err := os.Stat(dir + "/.vaultpull_schema.json")
	if err != nil {
		t.Fatalf("schema file not created: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected perm 0600, got %v", info.Mode().Perm())
	}
}

func TestLoadSchema_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.LoadSchema(dir)
	if err == nil {
		t.Fatal("expected error when schema file missing")
	}
}

func TestLoadSchema_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	sv := audit.SchemaVersion{Version: 2, Fields: []string{"key", "path"}}
	if err := audit.SaveSchema(dir, sv); err != nil {
		t.Fatalf("SaveSchema: %v", err)
	}
	loaded, err := audit.LoadSchema(dir)
	if err != nil {
		t.Fatalf("LoadSchema: %v", err)
	}
	if loaded.Version != 2 {
		t.Errorf("expected version 2, got %d", loaded.Version)
	}
	if len(loaded.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(loaded.Fields))
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestValidateEntries_Clean(t *testing.T) {
	sv := audit.SchemaVersion{Version: 1, Fields: []string{"key", "path", "action"}}
	entries := []audit.LogEntry{
		{Key: "DB_PASS", Path: "secret/app", Action: "added", Timestamp: time.Now()},
	}
	if err := audit.ValidateEntries(entries, sv); err != nil {
		t.Errorf("expected no issues, got: %v", err)
	}
}

func TestValidateEntries_DetectsMissingFields(t *testing.T) {
	sv := audit.SchemaVersion{Version: 1, Fields: []string{"key", "path", "action"}}
	entries := []audit.LogEntry{
		{Key: "", Path: "secret/app", Action: "added", Timestamp: time.Now()},
		{Key: "TOKEN", Path: "", Action: "updated", Timestamp: time.Now()},
	}
	verr := audit.ValidateEntries(entries, sv)
	if verr == nil {
		t.Fatal("expected validation error")
	}
	if len(verr.Issues) != 2 {
		t.Errorf("expected 2 issues, got %d: %v", len(verr.Issues), verr.Issues)
	}
}

func TestValidateEntries_EmptySchema(t *testing.T) {
	sv := audit.SchemaVersion{Version: 1, Fields: []string{}}
	entries := []audit.LogEntry{
		{Key: "", Path: "", Action: ""},
	}
	if err := audit.ValidateEntries(entries, sv); err != nil {
		t.Errorf("empty schema should pass all entries, got: %v", err)
	}
}
