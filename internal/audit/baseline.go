package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Baseline represents a saved state of secrets at a point in time.
type Baseline struct {
	Name      string            `json:"name"`
	CreatedAt time.Time         `json:"created_at"`
	Path      string            `json:"path"`
	Keys      map[string]string `json:"keys"`
}

func baselineFilePath(dir, name string) string {
	return filepath.Join(dir, fmt.Sprintf(".vaultpull_baseline_%s.json", name))
}

// SaveBaseline persists a baseline of the given secrets map under the provided name.
func SaveBaseline(dir, name, secretPath string, keys map[string]string) error {
	if name == "" {
		return fmt.Errorf("baseline name must not be empty")
	}
	b := Baseline{
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Path:      secretPath,
		Keys:      keys,
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal baseline: %w", err)
	}
	path := baselineFilePath(dir, name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write baseline: %w", err)
	}
	return nil
}

// LoadBaseline reads a previously saved baseline by name.
func LoadBaseline(dir, name string) (*Baseline, error) {
	if name == "" {
		return nil, fmt.Errorf("baseline name must not be empty")
	}
	path := baselineFilePath(dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("baseline %q not found", name)
		}
		return nil, fmt.Errorf("read baseline: %w", err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("parse baseline: %w", err)
	}
	return &b, nil
}
