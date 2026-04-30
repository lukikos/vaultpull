package audit

import (
	"os"
	"testing"
)

func TestSaveLabel_EmptyName(t *testing.T) {
	dir := t.TempDir()
	if err := SaveLabel(dir, "", "secret/app", "DB_PASS", "prod db"); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSaveLabel_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	if err := SaveLabel(dir, "my-label", "", "DB_PASS", ""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSaveLabel_EmptyKey(t *testing.T) {
	dir := t.TempDir()
	if err := SaveLabel(dir, "my-label", "secret/app", "", ""); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestSaveLabel_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := SaveLabel(dir, "prod", "secret/app", "API_KEY", "production api key"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(labelFilePath(dir)); err != nil {
		t.Fatalf("label file not created: %v", err)
	}
}

func TestSaveLabel_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	_ = SaveLabel(dir, "prod", "secret/app", "API_KEY", "")
	info, err := os.Stat(labelFilePath(dir))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

func TestLoadLabels_NoFile(t *testing.T) {
	dir := t.TempDir()
	labels, err := LoadLabels(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(labels) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(labels))
	}
}

func TestSaveLabel_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	_ = SaveLabel(dir, "prod", "secret/app", "API_KEY", "first note")
	_ = SaveLabel(dir, "prod", "secret/app", "API_KEY", "updated note")

	labels, err := LoadLabels(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(labels))
	}
	if labels[0].Note != "updated note" {
		t.Errorf("expected updated note, got %q", labels[0].Note)
	}
}

func TestSaveLabel_MultipleLabelsKeptSeparate(t *testing.T) {
	dir := t.TempDir()
	_ = SaveLabel(dir, "alpha", "secret/app", "DB_PASS", "alpha note")
	_ = SaveLabel(dir, "beta", "secret/app", "DB_PASS", "beta note")

	labels, err := LoadLabels(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}
}
