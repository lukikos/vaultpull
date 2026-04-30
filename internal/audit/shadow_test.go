package audit_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func TestSaveShadow_EmptyName(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveShadow(dir, "", []audit.ShadowEntry{{Path: "p", Key: "k", ValueHash: "h", RecordedAt: time.Now()}})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSaveShadow_EmptyEntries(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveShadow(dir, "test", nil)
	if err == nil {
		t.Fatal("expected error for empty entries")
	}
}

func TestSaveShadow_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	entries := []audit.ShadowEntry{
		{Path: "secret/app", Key: "DB_PASS", ValueHash: "abc123", RecordedAt: time.Now()},
	}
	if err := audit.SaveShadow(dir, "prod", entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(dir + "/.shadow_prod.json")
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected perm 0600, got %v", info.Mode().Perm())
	}
}

func TestLoadShadow_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.LoadShadow(dir, "missing")
	if err == nil {
		t.Fatal("expected error for missing shadow")
	}
}

func TestLoadShadow_EmptyName(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.LoadShadow(dir, "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestShadow_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	entries := []audit.ShadowEntry{
		{Path: "secret/app", Key: "API_KEY", ValueHash: "deadbeef", RecordedAt: now},
		{Path: "secret/app", Key: "DB_URL", ValueHash: "cafebabe", RecordedAt: now},
	}
	if err := audit.SaveShadow(dir, "staging", entries); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := audit.LoadShadow(dir, "staging")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded))
	}
	if loaded[0].Key != "API_KEY" || loaded[0].ValueHash != "deadbeef" {
		t.Errorf("unexpected entry: %+v", loaded[0])
	}
}

func TestCompareShadow_Match(t *testing.T) {
	shadow := []audit.ShadowEntry{
		{Path: "secret/app", Key: "TOKEN", ValueHash: "aaa"},
	}
	current := map[string]string{"TOKEN": "aaa"}
	reports := audit.CompareShadow("secret/app", current, shadow)
	if len(reports) != 1 || reports[0].Status != "match" {
		t.Errorf("expected match, got %+v", reports)
	}
}

func TestCompareShadow_Changed(t *testing.T) {
	shadow := []audit.ShadowEntry{
		{Path: "secret/app", Key: "TOKEN", ValueHash: "aaa"},
	}
	current := map[string]string{"TOKEN": "bbb"}
	reports := audit.CompareShadow("secret/app", current, shadow)
	if len(reports) != 1 || reports[0].Status != "changed" {
		t.Errorf("expected changed, got %+v", reports)
	}
}

func TestCompareShadow_Missing(t *testing.T) {
	shadow := []audit.ShadowEntry{
		{Path: "secret/app", Key: "GONE", ValueHash: "xyz"},
	}
	current := map[string]string{}
	reports := audit.CompareShadow("secret/app", current, shadow)
	if len(reports) != 1 || reports[0].Status != "missing" {
		t.Errorf("expected missing, got %+v", reports)
	}
}

func TestCompareShadow_FiltersByPath(t *testing.T) {
	shadow := []audit.ShadowEntry{
		{Path: "secret/other", Key: "X", ValueHash: "1"},
		{Path: "secret/app", Key: "Y", ValueHash: "2"},
	}
	current := map[string]string{"Y": "2"}
	reports := audit.CompareShadow("secret/app", current, shadow)
	if len(reports) != 1 || reports[0].Key != "Y" {
		t.Errorf("expected only Y, got %+v", reports)
	}
}
