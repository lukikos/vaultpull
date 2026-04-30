package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Label associates a human-readable name and optional description with a
// specific audit log entry identified by its timestamp and path.
type Label struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Key       string    `json:"key"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

func labelFilePath(dir string) string {
	return filepath.Join(dir, ".vaultpull_labels.json")
}

// SaveLabel persists a label entry to the labels file in dir.
func SaveLabel(dir, name, path, key, note string) error {
	if name == "" {
		return errors.New("label name must not be empty")
	}
	if path == "" {
		return errors.New("label path must not be empty")
	}
	if key == "" {
		return errors.New("label key must not be empty")
	}

	existing, _ := LoadLabels(dir)

	// Overwrite any existing label with the same name+path+key triple.
	updated := make([]Label, 0, len(existing)+1)
	for _, l := range existing {
		if l.Name == name && l.Path == path && l.Key == key {
			continue
		}
		updated = append(updated, l)
	}
	updated = append(updated, Label{
		Name:      name,
		Path:      path,
		Key:       key,
		Note:      note,
		CreatedAt: time.Now().UTC(),
	})

	data, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal labels: %w", err)
	}
	return os.WriteFile(labelFilePath(dir), data, 0o600)
}

// LoadLabels reads all labels from dir. Returns an empty slice when the file
// does not exist.
func LoadLabels(dir string) ([]Label, error) {
	data, err := os.ReadFile(labelFilePath(dir))
	if errors.Is(err, os.ErrNotExist) {
		return []Label{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read labels: %w", err)
	}
	var labels []Label
	if err := json.Unmarshal(data, &labels); err != nil {
		return nil, fmt.Errorf("parse labels: %w", err)
	}
	return labels, nil
}
