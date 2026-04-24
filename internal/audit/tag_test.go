package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveTag_EmptyName(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	err := SaveTag(logPath, "", "")
	if err == nil {
		t.Fatal("expected error for empty tag name")
	}
}

func TestSaveTag_CreatesTagFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	if err := SaveTag(logPath, "v1.0", "initial release"); err != nil {
		t.Fatalf("SaveTag: %v", err)
	}

	tagPath := tagFilePath(logPath)
	if _, err := os.Stat(tagPath); err != nil {
		t.Fatalf("tag file not created: %v", err)
	}
}

func TestSaveTag_AppendsTags(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	if err := SaveTag(logPath, "v1.0", "first"); err != nil {
		t.Fatalf("SaveTag first: %v", err)
	}
	if err := SaveTag(logPath, "v1.1", "second"); err != nil {
		t.Fatalf("SaveTag second: %v", err)
	}

	tags, err := LoadTags(logPath)
	if err != nil {
		t.Fatalf("LoadTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0].Name != "v1.0" || tags[1].Name != "v1.1" {
		t.Errorf("unexpected tag names: %+v", tags)
	}
}

func TestLoadTags_NoFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	tags, err := LoadTags(logPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tags != nil {
		t.Errorf("expected nil tags, got: %v", tags)
	}
}

func TestSaveTag_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	if err := SaveTag(logPath, "release", ""); err != nil {
		t.Fatalf("SaveTag: %v", err)
	}

	info, err := os.Stat(tagFilePath(logPath))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}
