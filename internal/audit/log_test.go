package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"testing"
)

func TestRecord_CreatesFile(t *testing.T) {
	tmp := t.TempDir() + "/audit.log"
	l := NewLogger(tmp)

	err := l.Record(Entry{
		SecretPath: "secret/app",
		Keys:       []string{"DB_PASS"},
		OutputFile: ".env",
		Status:     "success",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(tmp); os.IsNotExist(err) {
		t.Fatal("audit log file was not created")
	}
}

func TestRecord_AppendsEntries(t *testing.T) {
	tmp := t.TempDir() + "/audit.log"
	l := NewLogger(tmp)

	for i := 0; i < 3; i++ {
		if err := l.Record(Entry{SecretPath: "secret/app", Status: "success"}); err != nil {
			t.Fatalf("record %d failed: %v", i, err)
		}
	}

	f, _ := os.Open(tmp)
	defer f.Close()
	sc := bufio.NewScanner(f)
	count := 0
	for sc.Scan() {
		count++
	}
	if count != 3 {
		t.Fatalf("expected 3 lines, got %d", count)
	}
}

func TestRecord_ValidJSON(t *testing.T) {
	tmp := t.TempDir() + "/audit.log"
	l := NewLogger(tmp)
	l.Record(Entry{SecretPath: "secret/app", Status: "success", Error: "some err"})

	f, _ := os.Open(tmp)
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Scan()
	var e Entry
	if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if e.Status != "success" {
		t.Errorf("expected status success, got %s", e.Status)
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestRecord_FilePermissions(t *testing.T) {
	tmp := t.TempDir() + "/audit.log"
	l := NewLogger(tmp)
	l.Record(Entry{Status: "success"})

	info, err := os.Stat(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}
