package audit_test

import (
	"os"
	"testing"

	"github.com/your-org/vaultpull/internal/audit"
)

func TestSaveAnnotation_EmptyName(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveAnnotation(dir, "", "some note")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSaveAnnotation_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := audit.SaveAnnotation(dir, "v1.0", "initial release"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	annotations, err := audit.LoadAnnotations(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}
	if annotations[0].Name != "v1.0" {
		t.Errorf("expected name v1.0, got %q", annotations[0].Name)
	}
	if annotations[0].Note != "initial release" {
		t.Errorf("expected note 'initial release', got %q", annotations[0].Note)
	}
}

func TestSaveAnnotation_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SaveAnnotation(dir, "v1.0", "first note")
	_ = audit.SaveAnnotation(dir, "v1.0", "updated note")

	annotations, err := audit.LoadAnnotations(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation after overwrite, got %d", len(annotations))
	}
	if annotations[0].Note != "updated note" {
		t.Errorf("expected 'updated note', got %q", annotations[0].Note)
	}
}

func TestSaveAnnotation_MultipleNames(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SaveAnnotation(dir, "v1.0", "first")
	_ = audit.SaveAnnotation(dir, "v2.0", "second")

	annotations, err := audit.LoadAnnotations(dir)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(annotations))
	}
}

func TestLoadAnnotations_NoFile(t *testing.T) {
	dir := t.TempDir()
	annotations, err := audit.LoadAnnotations(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(annotations) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(annotations))
	}
}

func TestSaveAnnotation_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SaveAnnotation(dir, "v1.0", "note")

	info, err := os.Stat(dir + "/.vaultpull_annotations.json")
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("expected permissions 0600, got %o", perm)
	}
}
