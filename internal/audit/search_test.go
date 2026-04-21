package audit

import (
	"os"
	"testing"
	"time"
)

func makeSearchEntry(path, key, action string) Entry {
	return Entry{
		Timestamp: time.Now().UTC(),
		Path:      path,
		Key:       key,
		Action:    action,
	}
}

func TestSearch_NoFile(t *testing.T) {
	results, err := Search("/tmp/nonexistent_search_audit.log", SearchOptions{})
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_ByKeyContains(t *testing.T) {
	f, err := os.CreateTemp("", "search_audit_*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	logger, _ := NewLogger(f.Name())
	_ = logger.Record(makeSearchEntry("secret/app", "DB_PASSWORD", "added"))
	_ = logger.Record(makeSearchEntry("secret/app", "API_KEY", "added"))
	_ = logger.Record(makeSearchEntry("secret/app", "DB_HOST", "updated"))

	results, err := Search(f.Name(), SearchOptions{KeyContains: "db"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearch_ByPathContains(t *testing.T) {
	f, err := os.CreateTemp("", "search_audit_*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	logger, _ := NewLogger(f.Name())
	_ = logger.Record(makeSearchEntry("secret/app", "KEY", "added"))
	_ = logger.Record(makeSearchEntry("secret/infra", "TOKEN", "added"))

	results, err := Search(f.Name(), SearchOptions{PathContains: "infra"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "secret/infra" {
		t.Errorf("unexpected path: %s", results[0].Path)
	}
}

func TestSearch_ByActionIn(t *testing.T) {
	f, err := os.CreateTemp("", "search_audit_*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	logger, _ := NewLogger(f.Name())
	_ = logger.Record(makeSearchEntry("secret/app", "KEY1", "added"))
	_ = logger.Record(makeSearchEntry("secret/app", "KEY2", "updated"))
	_ = logger.Record(makeSearchEntry("secret/app", "KEY3", "unchanged"))

	results, err := Search(f.Name(), SearchOptions{ActionIn: []string{"added", "updated"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestSearch_CombinedCriteria(t *testing.T) {
	f, err := os.CreateTemp("", "search_audit_*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	logger, _ := NewLogger(f.Name())
	_ = logger.Record(makeSearchEntry("secret/app", "DB_PASSWORD", "added"))
	_ = logger.Record(makeSearchEntry("secret/app", "DB_HOST", "updated"))
	_ = logger.Record(makeSearchEntry("secret/infra", "DB_URL", "added"))

	results, err := Search(f.Name(), SearchOptions{
		KeyContains:  "db",
		PathContains: "app",
		ActionIn:     []string{"added"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].Key != "DB_PASSWORD" {
		t.Errorf("unexpected key: %s", results[0].Key)
	}
}
