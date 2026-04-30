package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MirrorEntry represents a point-in-time copy of a secret path's state.
type MirrorEntry struct {
	Name      string            `json:"name"`
	Path      string            `json:"path"`
	Keys      map[string]string `json:"keys"`
	CreatedAt time.Time         `json:"created_at"`
}

func mirrorFilePath(dir, name string) string {
	return filepath.Join(dir, fmt.Sprintf(".mirror_%s.json", name))
}

// SaveMirror persists a named mirror of the given secret path's current keys.
func SaveMirror(dir, name, path string, keys map[string]string) error {
	if name == "" {
		return fmt.Errorf("mirror name must not be empty")
	}
	if path == "" {
		return fmt.Errorf("mirror path must not be empty")
	}

	copy := make(map[string]string, len(keys))
	for k, v := range keys {
		copy[k] = v
	}

	entry := MirrorEntry{
		Name:      name,
		Path:      path,
		Keys:      copy,
		CreatedAt: time.Now().UTC(),
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mirror: %w", err)
	}

	fp := mirrorFilePath(dir, name)
	if err := os.WriteFile(fp, data, 0600); err != nil {
		return fmt.Errorf("write mirror: %w", err)
	}
	return nil
}

// LoadMirror reads a previously saved mirror by name.
func LoadMirror(dir, name string) (MirrorEntry, error) {
	if name == "" {
		return MirrorEntry{}, fmt.Errorf("mirror name must not be empty")
	}

	fp := mirrorFilePath(dir, name)
	data, err := os.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return MirrorEntry{}, fmt.Errorf("mirror %q not found", name)
		}
		return MirrorEntry{}, fmt.Errorf("read mirror: %w", err)
	}

	var entry MirrorEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return MirrorEntry{}, fmt.Errorf("parse mirror: %w", err)
	}
	return entry, nil
}
