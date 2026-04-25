package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ArchiveEntry represents a single archived audit log bundle.
type ArchiveEntry struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Entries   []Entry   `json:"entries"`
}

// archiveFilePath returns the path for a named archive file.
func archiveFilePath(dir, name string) string {
	return filepath.Join(dir, fmt.Sprintf(".vaultpull-archive-%s.json", name))
}

// Archive saves the current audit log entries under a named archive file.
// It does not remove the original log.
func Archive(logPath, dir, name string) error {
	if name == "" {
		return fmt.Errorf("archive name must not be empty")
	}

	entries, err := ReadAll(logPath)
	if err != nil {
		return fmt.Errorf("reading audit log: %w", err)
	}

	archive := ArchiveEntry{
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Entries:   entries,
	}

	data, err := json.MarshalIndent(archive, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling archive: %w", err)
	}

	dest := archiveFilePath(dir, name)
	if err := os.WriteFile(dest, data, 0600); err != nil {
		return fmt.Errorf("writing archive file: %w", err)
	}

	return nil
}

// LoadArchive reads a named archive file and returns its entry.
func LoadArchive(dir, name string) (*ArchiveEntry, error) {
	if name == "" {
		return nil, fmt.Errorf("archive name must not be empty")
	}

	path := archiveFilePath(dir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("archive %q not found", name)
		}
		return nil, fmt.Errorf("reading archive file: %w", err)
	}

	var archive ArchiveEntry
	if err := json.Unmarshal(data, &archive); err != nil {
		return nil, fmt.Errorf("parsing archive file: %w", err)
	}

	return &archive, nil
}
