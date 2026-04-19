package sync

import (
	"fmt"

	"github.com/user/vaultpull/internal/audit"
	"github.com/user/vaultpull/internal/config"
	"github.com/user/vaultpull/internal/dotenv"
	"github.com/user/vaultpull/internal/vault"
)

// Syncer orchestrates fetching secrets from Vault and writing them to a .env file.
type Syncer struct {
	cfg    *config.Config
	client *vault.Client
	audit  *audit.Logger
}

// New creates a Syncer from the provided config.
func New(cfg *config.Config) (*Syncer, error) {
	client, err := vault.NewClient(cfg.VaultAddr, cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("vault client: %w", err)
	}
	var al *audit.Logger
	if cfg.AuditLog != "" {
		al = audit.NewLogger(cfg.AuditLog)
	}
	return &Syncer{cfg: cfg, client: client, audit: al}, nil
}

// Run fetches secrets and writes them to the output .env file.
func (s *Syncer) Run() error {
	secrets, err := s.client.GetSecrets(s.cfg.SecretPath)

	entry := audit.Entry{
		SecretPath: s.cfg.SecretPath,
		OutputFile: s.cfg.OutputFile,
		Status:     "success",
	}

	if err != nil {
		entry.Status = "error"
		entry.Error = err.Error()
		s.recordAudit(entry)
		return fmt.Errorf("get secrets: %w", err)
	}

	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	entry.Keys = keys

	merged, err := dotenv.Merge(s.cfg.OutputFile, secrets)
	if err != nil {
		entry.Status = "error"
		entry.Error = err.Error()
		s.recordAudit(entry)
		return fmt.Errorf("merge: %w", err)
	}

	w := dotenv.NewWriter(s.cfg.OutputFile)
	if err := w.Write(merged); err != nil {
		entry.Status = "error"
		entry.Error = err.Error()
		s.recordAudit(entry)
		return fmt.Errorf("write: %w", err)
	}

	s.recordAudit(entry)
	return nil
}

func (s *Syncer) recordAudit(e audit.Entry) {
	if s.audit != nil {
		_ = s.audit.Record(e)
	}
}
