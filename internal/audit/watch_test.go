package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeWatchEntries(t *testing.T, logFile string, entries []Entry) {
	t.Helper()
	l, err := NewLogger(logFile)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for _, e := range entries {
		if err := l.Record(e.Path, e.Key, e.Action); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}
}

func TestWatch_NoNewEntries(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")

	done := make(chan struct{})
	opts := WatchOptions{
		LogFile:  logFile,
		Interval: 10 * time.Millisecond,
		MaxPolls: 2,
	}
	events, err := Watch(opts, done)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	var got []WatchEvent
	for e := range events {
		got = append(got, e)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 events, got %d", len(got))
	}
}

func TestWatch_EmitsNewEntries(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")

	writeWatchEntries(t, logFile, []Entry{
		{Path: "secret/app", Key: "DB_PASS", Action: "added"},
		{Path: "secret/app", Key: "API_KEY", Action: "updated"},
	})

	done := make(chan struct{})
	opts := WatchOptions{
		LogFile:  logFile,
		Interval: 10 * time.Millisecond,
		MaxPolls: 1,
	}
	events, err := Watch(opts, done)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	var got []WatchEvent
	for e := range events {
		got = append(got, e)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].Key != "DB_PASS" {
		t.Errorf("expected first key DB_PASS, got %s", got[0].Key)
	}
}

func TestWatch_FiltersByPath(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")

	writeWatchEntries(t, logFile, []Entry{
		{Path: "secret/app", Key: "DB_PASS", Action: "added"},
		{Path: "secret/other", Key: "TOKEN", Action: "added"},
	})

	done := make(chan struct{})
	opts := WatchOptions{
		LogFile:  logFile,
		Path:     "secret/app",
		Interval: 10 * time.Millisecond,
		MaxPolls: 1,
	}
	events, err := Watch(opts, done)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	var got []WatchEvent
	for e := range events {
		got = append(got, e)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	if got[0].Path != "secret/app" {
		t.Errorf("unexpected path %s", got[0].Path)
	}
}

func TestWatch_DoneStops(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")
	_ = os.WriteFile(logFile, []byte{}, 0o600)

	done := make(chan struct{})
	opts := WatchOptions{
		LogFile:  logFile,
		Interval: 20 * time.Millisecond,
	}
	events, err := Watch(opts, done)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	close(done)
	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case _, ok := <-events:
			if !ok {
				return // channel closed as expected
			}
		case <-timeout:
			t.Fatal("watcher did not stop after done was closed")
		}
	}
}
