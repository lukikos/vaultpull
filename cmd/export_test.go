package cmd

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"vaultpull/internal/audit"
)

func TestRunExport_NoEntries(t *testing.T) {
	tmp := t.TempDir()
	auditLog = filepath.Join(tmp, "audit.jsonl")

	var buf bytes.Buffer
	exportCmd.SetOut(&buf)
	exportOutput = filepath.Join(tmp, "out.csv")

	if err := runExport(exportCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "No audit entries") {
		t.Errorf("expected no-entries message, got: %s", buf.String())
	}
}

func TestRunExport_WritesCSV(t *testing.T) {
	tmp := t.TempDir()
	auditLog = filepath.Join(tmp, "audit.jsonl")

	logger, err := audit.NewLogger(auditLog)
	if err != nil {
		t.Fatalf("creating logger: %v", err)
	}
	_ = logger.Record(audit.Entry{
		Timestamp: time.Now(),
		SecretPath: "secret/data/app",
		Keys:       []string{"DB_PASS"},
		Status:     "success",
	})

	outFile := filepath.Join(tmp, "out.csv")
	exportOutput = outFile

	var buf bytes.Buffer
	exportCmd.SetOut(&buf)

	if err := runExport(exportCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("reading csv: %v", err)
	}

	r := csv.NewReader(bytes.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parsing csv: %v", err)
	}

	if len(records) < 2 {
		t.Fatalf("expected header + 1 row, got %d rows", len(records))
	}

	if records[0][0] != "timestamp" {
		t.Errorf("expected header row, got: %v", records[0])
	}
}
