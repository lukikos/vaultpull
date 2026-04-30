package audit

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// Checkpoint records a named point-in-time marker tied to a specific audit log offset.
type Checkpoint struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Offset    int       `json:"offset"`
	CreatedAt time.Time `json:"created_at"`
}

func checkpointFilePath(dir string) string {
	return filepath.Join(dir, ".vaultpull_checkpoints.json")
}

// SaveCheckpoint persists a named checkpoint for the given vault path and log offset.
func SaveCheckpoint(dir, name, path string, offset int) error {
	if name == "" {
		return errors.New("checkpoint name must not be empty")
	}
	if path == "" {
		return errors.New("checkpoint path must not be empty")
	}

	existing, _ := LoadCheckpoints(dir)

	updated := make([]Checkpoint, 0, len(existing)+1)
	for _, c := range existing {
		if c.Name != name || c.Path != path {
			updated = append(updated, c)
		}
	}
	updated = append(updated, Checkpoint{
		Name:      name,
		Path:      path,
		Offset:    offset,
		CreatedAt: time.Now().UTC(),
	})

	data, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(checkpointFilePath(dir), data, 0o600)
}

// LoadCheckpoints reads all saved checkpoints from dir.
func LoadCheckpoints(dir string) ([]Checkpoint, error) {
	data, err := os.ReadFile(checkpointFilePath(dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []Checkpoint
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// FindCheckpoint returns the checkpoint matching name and path, or nil if not found.
func FindCheckpoint(dir, name, path string) (*Checkpoint, error) {
	all, err := LoadCheckpoints(dir)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].Name == name && all[i].Path == path {
			return &all[i], nil
		}
	}
	return nil, nil
}
