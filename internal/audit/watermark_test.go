package audit_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func makeWatermarkEntry(path, action string, ts time.Time) audit.Entry {
	return audit.Entry{Path: path, Action: action, Timestamp: ts}
}

func TestComputeWatermarks_Empty(t *testing.T) {
	wm := audit.ComputeWatermarks(nil)
	if len(wm) != 0 {
		t.Fatalf("expected 0 watermarks, got %d", len(wm))
	}
}

func TestComputeWatermarks_IgnoresNonSync(t *testing.T) {
	entries := []audit.Entry{
		makeWatermarkEntry("secret/app", "read", time.Now()),
	}
	wm := audit.ComputeWatermarks(entries)
	if len(wm) != 0 {
		t.Fatalf("expected 0 watermarks for non-sync entries, got %d", len(wm))
	}
}

func TestComputeWatermarks_SinglePath(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	entries := []audit.Entry{
		makeWatermarkEntry("secret/app", "sync", now.Add(-2*time.Hour)),
		makeWatermarkEntry("secret/app", "sync", now),
	}
	wm := audit.ComputeWatermarks(entries)
	if len(wm) != 1 {
		t.Fatalf("expected 1 watermark, got %d", len(wm))
	}
	if !wm[0].LatestAt.Equal(now) {
		t.Errorf("expected latest %v, got %v", now, wm[0].LatestAt)
	}
	if wm[0].SyncCount != 2 {
		t.Errorf("expected sync count 2, got %d", wm[0].SyncCount)
	}
}

func TestComputeWatermarks_MultiplePaths(t *testing.T) {
	now := time.Now().UTC()
	entries := []audit.Entry{
		makeWatermarkEntry("secret/a", "sync", now),
		makeWatermarkEntry("secret/b", "sync", now.Add(-time.Hour)),
	}
	wm := audit.ComputeWatermarks(entries)
	if len(wm) != 2 {
		t.Fatalf("expected 2 watermarks, got %d", len(wm))
	}
}

func TestSaveAndLoadWatermarks_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	entries := []audit.Entry{
		makeWatermarkEntry("secret/app", "sync", now),
	}
	wm := audit.ComputeWatermarks(entries)
	if err := audit.SaveWatermarks(dir, wm); err != nil {
		t.Fatalf("SaveWatermarks: %v", err)
	}
	loaded, err := audit.LoadWatermarks(dir)
	if err != nil {
		t.Fatalf("LoadWatermarks: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 watermark, got %d", len(loaded))
	}
	if !loaded[0].LatestAt.Equal(now) {
		t.Errorf("timestamp mismatch: want %v got %v", now, loaded[0].LatestAt)
	}
}

func TestLoadWatermarks_NotFound(t *testing.T) {
	dir := t.TempDir()
	wm, err := audit.LoadWatermarks(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wm != nil {
		t.Errorf("expected nil, got %v", wm)
	}
}

func TestSaveWatermarks_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	if err := audit.SaveWatermarks(dir, []audit.Watermark{}); err != nil {
		t.Fatalf("SaveWatermarks: %v", err)
	}
	path := filepath.Join(dir, ".vaultpull_watermarks.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("expected 0600, got %o", perm)
	}
	// Validate JSON is well-formed
	data, _ := os.ReadFile(path)
	var out []audit.Watermark
	if err := json.Unmarshal(data, &out); err != nil {
		t.Errorf("invalid JSON: %v", err)
	}
}
