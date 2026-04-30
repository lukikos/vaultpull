package audit_test

import (
	"os"
	"testing"

	"github.com/yourusername/vaultpull/internal/audit"
)

func TestSaveMirror_EmptyName(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveMirror(dir, "", "secret/app", map[string]string{"KEY": "val"})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSaveMirror_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveMirror(dir, "prod", "", map[string]string{"KEY": "val"})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSaveMirror_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	keys := map[string]string{"DB_PASS": "secret", "API_KEY": "abc123"}

	if err := audit.SaveMirror(dir, "baseline", "secret/app", keys); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry, err := audit.LoadMirror(dir, "baseline")
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if entry.Name != "baseline" {
		t.Errorf("expected name %q, got %q", "baseline", entry.Name)
	}
	if entry.Path != "secret/app" {
		t.Errorf("expected path %q, got %q", "secret/app", entry.Path)
	}
	if entry.Keys["DB_PASS"] != "secret" {
		t.Errorf("expected DB_PASS=secret, got %q", entry.Keys["DB_PASS"])
	}
	if entry.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestSaveMirror_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	if err := audit.SaveMirror(dir, "perm", "secret/x", map[string]string{"K": "v"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(dir + "/.mirror_perm.json")
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %v", info.Mode().Perm())
	}
}

func TestLoadMirror_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.LoadMirror(dir, "missing")
	if err == nil {
		t.Fatal("expected error for missing mirror")
	}
}

func TestLoadMirror_EmptyName(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.LoadMirror(dir, "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSaveMirror_DoesNotMutateOriginal(t *testing.T) {
	dir := t.TempDir()
	original := map[string]string{"KEY": "original"}
	if err := audit.SaveMirror(dir, "snap", "secret/app", original); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	original["KEY"] = "mutated"

	entry, err := audit.LoadMirror(dir, "snap")
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if entry.Keys["KEY"] != "original" {
		t.Errorf("mirror was mutated: got %q", entry.Keys["KEY"])
	}
}
