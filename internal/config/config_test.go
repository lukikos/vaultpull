package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_MissingToken(t *testing.T) {
	os.Unsetenv("VAULT_TOKEN")
	os.Unsetenv("VAULTPULL_VAULT_TOKEN")

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error when vault token is missing")
	}
}

func TestLoad_MissingSecretPath(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "test-token")

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error when secret_path is missing")
	}
}

func TestLoad_FromConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := []byte(`
vault_addr: http://vault.example.com:8200
vault_token: s.testtoken
secret_path: myapp/prod
output_file: prod.env
mount_path: kv
`)
	if err := os.WriteFile(cfgPath, content, 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.VaultAddr != "http://vault.example.com:8200" {
		t.Errorf("unexpected vault_addr: %s", cfg.VaultAddr)
	}
	if cfg.VaultToken != "s.testtoken" {
		t.Errorf("unexpected vault_token: %s", cfg.VaultToken)
	}
	if cfg.SecretPath != "myapp/prod" {
		t.Errorf("unexpected secret_path: %s", cfg.SecretPath)
	}
	if cfg.OutputFile != "prod.env" {
		t.Errorf("unexpected output_file: %s", cfg.OutputFile)
	}
	if cfg.MountPath != "kv" {
		t.Errorf("unexpected mount_path: %s", cfg.MountPath)
	}
}

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := []byte(`
vault_token: s.testtoken
secret_path: myapp/dev
`)
	if err := os.WriteFile(cfgPath, content, 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.VaultAddr != "http://127.0.0.1:8200" {
		t.Errorf("expected default vault_addr, got: %s", cfg.VaultAddr)
	}
	if cfg.OutputFile != ".env" {
		t.Errorf("expected default output_file, got: %s", cfg.OutputFile)
	}
	if cfg.MountPath != "secret" {
		t.Errorf("expected default mount_path, got: %s", cfg.MountPath)
	}
}
