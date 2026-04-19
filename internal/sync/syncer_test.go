package sync_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/vaultpull/internal/config"
	"github.com/example/vaultpull/internal/sync"
)

func newFakeVault(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"data":{"API_KEY":"abc123","DB_PASS":"secret"}}}`))
	}))
}

func TestRun_WritesSecrets(t *testing.T) {
	srv := newFakeVault(t)
	defer srv.Close()

	tmpDir := t.TempDir()
	output := filepath.Join(tmpDir, ".env")

	cfg := &config.Config{
		VaultAddress: srv.URL,
		VaultToken:   "test-token",
		SecretPath:   "secret/myapp",
		OutputFile:   output,
		Overwrite:    true,
		Backup:       false,
	}

	s, err := sync.New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := s.Run()
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if result.SecretsCount != 2 {
		t.Errorf("expected 2 secrets, got %d", result.SecretsCount)
	}
	if result.OutputPath != output {
		t.Errorf("unexpected output path: %s", result.OutputPath)
	}
	if result.BackedUp {
		t.Error("expected no backup")
	}

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	if len(data) == 0 {
		t.Error("output file is empty")
	}
}

func TestNew_InvalidAddress(t *testing.T) {
	cfg := &config.Config{
		VaultAddress: "://bad-url",
		VaultToken:   "token",
		SecretPath:   "secret/app",
		OutputFile:   ".env",
	}
	_, err := sync.New(cfg)
	if err == nil {
		t.Error("expected error for invalid address")
	}
}
