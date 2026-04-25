package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Snapshot represents a named point-in-time capture of secrets for a path.
type Snapshot struct {
	Name      string            `json:"name"`
	Path      string            `json:"path"`
	CreatedAt time.Time         `json:"created_at"`
	Secrets   map[string]string `json:"secrets"`
}

func snapshotFilePath(dir, name string) string {
	return filepath.Join(dir, fmt.Sprintf(".vaultpull.snapshot.%s.json", name))
}

// SaveSnapshot persists a named snapshot of secrets to the given directory.
func SaveSnapshot(dir, name, path string, secrets map[string]string) error {
	if name == "" {
		return fmt.Errorf("snapshot name must not be empty")
	}
	s := Snapshot{
		Name:      name,
		Path:      path,
		CreatedAt: time.Now().UTC(),
		Secrets:   secrets,
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	dest := snapshotFilePath(dir, name)
	if err := os.WriteFile(dest, data, 0600); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}
	return nil
}

// LoadSnapshot reads a named snapshot from the given directory.
func LoadSnapshot(dir, name string) (*Snapshot, error) {
	if name == "" {
		return nil, fmt.Errorf("snapshot name must not be empty")
	}
	src := snapshotFilePath(dir, name)
	data, err := os.ReadFile(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("snapshot %q not found", name)
		}
		return nil, fmt.Errorf("read snapshot: %w", err)
	}
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse snapshot: %w", err)
	}
	return &s, nil
}
