package sync

import (
	"fmt"

	"github.com/example/vaultpull/internal/config"
	"github.com/example/vaultpull/internal/dotenv"
	"github.com/example/vaultpull/internal/vault"
)

// Result holds the outcome of a sync operation.
type Result struct {
	SecretsCount int
	OutputPath   string
	BackedUp     bool
}

// Syncer orchestrates pulling secrets from Vault and writing them to a .env file.
type Syncer struct {
	client *vault.Client
	cfg    *config.Config
}

// New creates a new Syncer from the provided config.
func New(cfg *config.Config) (*Syncer, error) {
	client, err := vault.NewClient(cfg.VaultAddress, cfg.VaultToken)
	if err != nil {
		return nil, fmt.Errorf("sync: failed to create vault client: %w", err)
	}
	return &Syncer{client: client, cfg: cfg}, nil
}

// Run performs the full sync: fetch secrets, merge with existing file, write output.
func (s *Syncer) Run() (*Result, error) {
	secrets, err := s.client.GetSecrets(s.cfg.SecretPath)
	if err != nil {
		return nil, fmt.Errorf("sync: failed to get secrets: %w", err)
	}

	merged, err := dotenv.Merge(s.cfg.OutputFile, secrets, s.cfg.Overwrite)
	if err != nil {
		return nil, fmt.Errorf("sync: failed to merge secrets: %w", err)
	}

	w, err := dotenv.NewWriter(s.cfg.OutputFile, s.cfg.Backup)
	if err != nil {
		return nil, fmt.Errorf("sync: failed to create writer: %w", err)
	}

	bacedUp, err := w.Write(merged)
	if err != nil {
		return nil, fmt.Errorf("sync: failed to write output: %w", err)
	}

	return &Result{
		SecretsCount: len(merged),
		OutputPath:   s.cfg.OutputFile,
		BackedUp:     bacedUp,
	}, nil
}
