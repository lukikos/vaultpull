package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMerge_NoExistingFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")

	result, err := Merge(p, map[string]string{"KEY": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["KEY"] != "val" {
		t.Errorf("expected val, got %q", result["KEY"])
	}
}

func TestMerge_PreservesExistingKeys(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	_ = os.WriteFile(p, []byte("EXISTING=keep\nOLD=old\n"), 0600)

	result, err := Merge(p, map[string]string{"NEW": "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["EXISTING"] != "keep" {
		t.Errorf("EXISTING should be preserved, got %q", result["EXISTING"])
	}
	if result["NEW"] != "new" {
		t.Errorf("NEW should be set, got %q", result["NEW"])
	}
}

func TestMerge_OverwritesExistingKey(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	_ = os.WriteFile(p, []byte("TOKEN=old_token\n"), 0600)

	result, err := Merge(p, map[string]string{"TOKEN": "new_token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["TOKEN"] != "new_token" {
		t.Errorf("TOKEN should be overwritten, got %q", result["TOKEN"])
	}
}

func TestParse_IgnoresComments(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	_ = os.WriteFile(p, []byte("# comment\nKEY=value\n\nANOTHER=one\n"), 0600)

	result, err := parse(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result))
	}
}
