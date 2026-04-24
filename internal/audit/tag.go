package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Tag represents a named marker attached to a point in the audit log.
type Tag struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Note      string    `json:"note,omitempty"`
}

// SaveTag writes a tag entry to a sidecar JSON file alongside the audit log.
func SaveTag(logPath, name, note string) error {
	if name == "" {
		return fmt.Errorf("tag name must not be empty")
	}
	tagPath := tagFilePath(logPath)

	tags, err := LoadTags(logPath)
	if err != nil {
		tags = []Tag{}
	}

	tags = append(tags, Tag{
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Note:      note,
	})

	f, err := os.OpenFile(tagPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open tag file: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(tags)
}

// LoadTags reads all tags from the sidecar tag file.
func LoadTags(logPath string) ([]Tag, error) {
	tagPath := tagFilePath(logPath)
	f, err := os.Open(tagPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open tag file: %w", err)
	}
	defer f.Close()

	var tags []Tag
	if err := json.NewDecoder(f).Decode(&tags); err != nil {
		return nil, fmt.Errorf("decode tags: %w", err)
	}
	return tags, nil
}

func tagFilePath(logPath string) string {
	return logPath + ".tags.json"
}
