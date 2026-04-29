package audit

import (
	"testing"
	"time"
)

func TestSavePin_EmptyName(t *testing.T) {
	dir := t.TempDir()
	err := SavePin(dir, "", "secret/app", "DB_PASS", "hunter2", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSavePin_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	err := SavePin(dir, "my-pin", "", "DB_PASS", "hunter2", "")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSavePin_EmptyKey(t *testing.T) {
	dir := t.TempDir()
	err := SavePin(dir, "my-pin", "secret/app", "", "hunter2", "")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestSavePin_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := SavePin(dir, "release-v1", "secret/app", "API_KEY", "abc123", "pinned at release"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pins, err := LoadPins(dir)
	if err != nil {
		t.Fatalf("LoadPins error: %v", err)
	}
	if len(pins) != 1 {
		t.Fatalf("expected 1 pin, got %d", len(pins))
	}
	p := pins[0]
	if p.Name != "release-v1" {
		t.Errorf("expected name 'release-v1', got %q", p.Name)
	}
	if p.Key != "API_KEY" {
		t.Errorf("expected key 'API_KEY', got %q", p.Key)
	}
	if p.Value != "abc123" {
		t.Errorf("expected value 'abc123', got %q", p.Value)
	}
	if p.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestSavePin_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	_ = SavePin(dir, "pin-a", "secret/app", "KEY", "old", "")
	_ = SavePin(dir, "pin-a", "secret/app", "KEY", "new", "updated")

	pins, _ := LoadPins(dir)
	if len(pins) != 1 {
		t.Fatalf("expected 1 pin after overwrite, got %d", len(pins))
	}
	if pins[0].Value != "new" {
		t.Errorf("expected value 'new', got %q", pins[0].Value)
	}
}

func TestSavePin_MultiplePins(t *testing.T) {
	dir := t.TempDir()
	_ = SavePin(dir, "alpha", "secret/app", "KEY_A", "val-a", "")
	_ = SavePin(dir, "beta", "secret/app", "KEY_B", "val-b", "")

	pins, _ := LoadPins(dir)
	if len(pins) != 2 {
		t.Fatalf("expected 2 pins, got %d", len(pins))
	}
}

func TestLoadPins_NoFile(t *testing.T) {
	dir := t.TempDir()
	pins, err := LoadPins(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pins) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(pins))
	}
}

func TestSavePin_CreatedAtIsRecent(t *testing.T) {
	dir := t.TempDir()
	before := time.Now().UTC().Add(-time.Second)
	_ = SavePin(dir, "ts-pin", "secret/app", "KEY", "val", "")
	after := time.Now().UTC().Add(time.Second)

	pins, _ := LoadPins(dir)
	if len(pins) == 0 {
		t.Fatal("expected pin to be saved")
	}
	cat := pins[0].CreatedAt
	if cat.Before(before) || cat.After(after) {
		t.Errorf("CreatedAt %v not within expected range [%v, %v]", cat, before, after)
	}
}
