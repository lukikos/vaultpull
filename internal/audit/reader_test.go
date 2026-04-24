package audit

import (
	"testing"
)

func TestReadAll_EmptyWhenNoFile(t *testing.T) {
	entries, err := ReadAll("/nonexistent/audit.log")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(entries))
	}
}

func TestReadAll_ReturnsWrittenEntries(t *testing.T) {
	tmp := t.TempDir() + "/audit.log"
	l := NewLogger(tmp)

	expected := []Entry{
		{SecretPath: "secret/app", Keys: []string{"KEY1"}, OutputFile: ".env", Status: "success"},
		{SecretPath: "secret/db", Keys: []string{"DB_PASS"}, OutputFile: ".env.db", Status: "error", Error: "not found"},
	}
	for _, e := range expected {
		if err := l.Record(e); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := ReadAll(tmp)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].SecretPath != "secret/app" {
		t.Errorf("unexpected secret path: %s", entries[0].SecretPath)
	}
	if entries[1].Error != "not found" {
		t.Errorf("unexpected error field: %s", entries[1].Error)
	}
}

func TestReadAll_PreservesEntryOrder(t *testing.T) {
	tmp := t.TempDir() + "/audit.log"
	l := NewLogger(tmp)

	paths := []string{"secret/first", "secret/second", "secret/third"}
	for _, p := range paths {
		if err := l.Record(Entry{SecretPath: p, Status: "success"}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := ReadAll(tmp)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if len(entries) != len(paths) {
		t.Fatalf("expected %d entries, got %d", len(paths), len(entries))
	}
	for i, p := range paths {
		if entries[i].SecretPath != p {
			t.Errorf("entry %d: expected path %q, got %q", i, p, entries[i].SecretPath)
		}
	}
}
