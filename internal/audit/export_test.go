package audit

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"
)

func TestExportCSV_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := ExportCSV(nil, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (header), got %d", len(lines))
	}
}

func TestExportCSV_Header(t *testing.T) {
	var buf bytes.Buffer
	_ = ExportCSV(nil, &buf)
	r := csv.NewReader(&buf)
	header, err := r.Read()
	if err != nil {
		t.Fatalf("reading header: %v", err)
	}
	expected := []string{"timestamp", "path", "keys_written", "output_file", "status"}
	for i, col := range expected {
		if header[i] != col {
			t.Errorf("col %d: want %q got %q", i, col, header[i])
		}
	}
}

func TestExportCSV_Rows(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	entries := []Entry{
		{Timestamp: now, Path: "secret/app", KeysWritten: 3, OutputFile: ".env", Status: "ok"},
		{Timestamp: now, Path: "secret/db", KeysWritten: 1, OutputFile: ".env.db", Status: "ok"},
	}

	var buf bytes.Buffer
	if err := ExportCSV(entries, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := csv.NewReader(&buf)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("reading csv: %v", err)
	}
	// header + 2 rows
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
	if records[1][1] != "secret/app" {
		t.Errorf("row 1 path: want %q got %q", "secret/app", records[1][1])
	}
	if records[2][2] != "1" {
		t.Errorf("row 2 keys_written: want %q got %q", "1", records[2][2])
	}
}
