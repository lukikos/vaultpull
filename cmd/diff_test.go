package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDiff_NoExistingFile(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, ".env")

	// Point config at a non-existent output file — all keys should show as added.
	t.Setenv("VAULT_TOKEN", "test-token")
	t.Setenv("VAULT_SECRET_PATH", "secret/data/myapp")
	t.Setenv("VAULTPULL_OUTPUT", outFile)

	// Provide a fake Vault server via the test helper used in syncer_test.
	srv := newFakeVaultHTTP(t, map[string]string{"API_KEY": "abc"})
	t.Setenv("VAULT_ADDR", srv.URL)

	var buf bytes.Buffer
	diffCmd.SetOut(&buf)

	err := diffCmd.RunE(diffCmd, nil)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "+ API_KEY")
	assert.Contains(t, out, "1 added")
}

func TestRunDiff_DetectsUpdate(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, ".env")

	// Write an existing .env with a stale value.
	require.NoError(t, os.WriteFile(outFile, []byte("API_KEY=old\n"), 0600))

	t.Setenv("VAULT_TOKEN", "test-token")
	t.Setenv("VAULT_SECRET_PATH", "secret/data/myapp")
	t.Setenv("VAULTPULL_OUTPUT", outFile)

	srv := newFakeVaultHTTP(t, map[string]string{"API_KEY": "new"})
	t.Setenv("VAULT_ADDR", srv.URL)

	var buf bytes.Buffer
	diffCmd.SetOut(&buf)

	err := diffCmd.RunE(diffCmd, nil)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "~ API_KEY")
	assert.Contains(t, out, "1 updated")
}
