package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultpull/internal/audit"
)

func newSchemaCmd(t *testing.T, dir string) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	saveCmd := &cobra.Command{Use: "save", RunE: runSchemaSave}
	saveCmd.Flags().IntP("version", "v", 1, "")
	saveCmd.Flags().StringSlice("fields", []string{"key", "path", "action"}, "")
	saveCmd.Flags().String("dir", dir, "")
	saveCmd.SetOut(buf)

	validateCmd := &cobra.Command{Use: "validate", RunE: runSchemaValidate}
	validateCmd.Flags().String("log", filepath.Join(dir, "audit.log"), "")
	validateCmd.Flags().String("dir", dir, "")
	validateCmd.SetOut(buf)

	root := &cobra.Command{Use: "schema"}
	root.AddCommand(saveCmd, validateCmd)
	root.SetOut(buf)
	return root, buf
}

func TestRunSchemaSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	cmd, buf := newSchemaCmd(t, dir)
	cmd.SetArgs([]string{"save", "--version", "1", "--dir", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Schema v1 saved") {
		t.Errorf("expected confirmation, got: %s", buf.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".vaultpull_schema.json")); err != nil {
		t.Errorf("schema file not created: %v", err)
	}
}

func TestRunSchemaValidate_OK(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "audit.log")

	// save schema
	if err := audit.SaveSchema(dir, audit.SchemaVersion{Version: 1, Fields: []string{"key", "path", "action"}}); err != nil {
		t.Fatalf("SaveSchema: %v", err)
	}

	// write a valid log entry
	logger, err := audit.NewLogger(logFile)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	if err := logger.Record("secret/app", "DB_PASS", "added"); err != nil {
		t.Fatalf("Record: %v", err)
	}

	cmd, buf := newSchemaCmd(t, dir)
	cmd.SetArgs([]string{"validate", "--log", logFile, "--dir", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "OK") {
		t.Errorf("expected OK output, got: %s", buf.String())
	}
}

func TestRunSchemaValidate_MissingSchema(t *testing.T) {
	dir := t.TempDir()
	cmd, _ := newSchemaCmd(t, dir)
	cmd.SetArgs([]string{"validate", "--log", filepath.Join(dir, "audit.log"), "--dir", dir})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when schema file is missing")
	}
}
