package audit

import (
	"os"
	"testing"
)

func TestSaveSnapshot_EmptyName(t *testing.T) {
	dir := t.TempDir()
	err := SaveSnapshot(dir, "", "secret/app", map[string]string{"KEY": "val"})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSaveSnapshot_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	secrets := map[string]string{"DB_PASS": "hunter2", "API_KEY": "abc123"}
	if err := SaveSnapshot(dir, "v1", "secret/app", secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	path := snapshotFilePath(dir, "v1")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("snapshot file not created: %v", err)
	}
}

func TestSaveSnapshot_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	if err := SaveSnapshot(dir, "perm-test", "secret/app", map[string]string{"X": "y"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(snapshotFilePath(dir, "perm-test"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

func TestLoadSnapshot_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadSnapshot(dir, "missing")
	if err == nil {
		t.Fatal("expected error for missing snapshot")
	}
}

func TestLoadSnapshot_EmptyName(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadSnapshot(dir, "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestLoadSnapshot_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	secrets := map[string]string{"TOKEN": "secret", "PORT": "5432"}
	if err := SaveSnapshot(dir, "release", "secret/myapp", secrets); err != nil {
		t.Fatalf("save: %v", err)
	}
	s, err := LoadSnapshot(dir, "release")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if s.Name != "release" {
		t.Errorf("name: got %q, want %q", s.Name, "release")
	}
	if s.Path != "secret/myapp" {
		t.Errorf("path: got %q, want %q", s.Path, "secret/myapp")
	}
	if len(s.Secrets) != len(secrets) {
		t.Errorf("secrets count: got %d, want %d", len(s.Secrets), len(secrets))
	}
	for k, v := range secrets {
		if s.Secrets[k] != v {
			t.Errorf("key %q: got %q, want %q", k, s.Secrets[k], v)
		}
	}
}
