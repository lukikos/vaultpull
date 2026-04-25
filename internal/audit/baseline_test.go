package audit_test

import (
	"os"
	"testing"

	"github.com/yourusername/vaultpull/internal/audit"
)

func TestSaveBaseline_EmptyName(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveBaseline(dir, "", "secret/app", map[string]string{"KEY": "val"})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSaveBaseline_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	keys := map[string]string{"DB_PASS": "secret", "API_KEY": "abc123"}
	if err := audit.SaveBaseline(dir, "v1", "secret/app", keys); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
}

func TestSaveBaseline_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	if err := audit.SaveBaseline(dir, "prod", "secret/app", map[string]string{"X": "y"}); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	info, err := entries[0].Info()
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

func TestLoadBaseline_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.LoadBaseline(dir, "missing")
	if err == nil {
		t.Fatal("expected error for missing baseline")
	}
}

func TestLoadBaseline_EmptyName(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.LoadBaseline(dir, "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestLoadBaseline_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	keys := map[string]string{"TOKEN": "abc", "HOST": "localhost"}
	if err := audit.SaveBaseline(dir, "staging", "secret/staging", keys); err != nil {
		t.Fatal(err)
	}
	b, err := audit.LoadBaseline(dir, "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Name != "staging" {
		t.Errorf("expected name staging, got %q", b.Name)
	}
	if b.Path != "secret/staging" {
		t.Errorf("expected path secret/staging, got %q", b.Path)
	}
	if len(b.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(b.Keys))
	}
	if b.Keys["TOKEN"] != "abc" {
		t.Errorf("expected TOKEN=abc, got %q", b.Keys["TOKEN"])
	}
}
