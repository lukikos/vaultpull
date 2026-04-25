package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// RollbackResult describes the outcome of a rollback operation.
type RollbackResult struct {
	TagName   string
	Path      string
	Restored  map[string]string
	Timestamp time.Time
}

// Rollback restores the secret state for a given path to the snapshot
// captured at the named tag. It returns the key/value pairs that were
// restored so the caller can write them to a .env file.
func Rollback(dir, tagName, path string) (*RollbackResult, error) {
	if tagName == "" {
		return nil, fmt.Errorf("tag name must not be empty")
	}
	if path == "" {
		return nil, fmt.Errorf("path must not be empty")
	}

	tags, err := LoadTags(dir)
	if err != nil {
		return nil, fmt.Errorf("load tags: %w", err)
	}

	var target *Tag
	for i := range tags {
		if tags[i].Name == tagName {
			target = &tags[i]
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("tag %q not found", tagName)
	}

	snap, err := LoadSnapshot(dir, tagName)
	if err != nil {
		return nil, fmt.Errorf("load snapshot for tag %q: %w", tagName, err)
	}

	state, ok := snap[path]
	if !ok {
		return nil, fmt.Errorf("no snapshot data for path %q in tag %q", path, tagName)
	}

	return &RollbackResult{
		TagName:   tagName,
		Path:      path,
		Restored:  state,
		Timestamp: target.CreatedAt,
	}, nil
}

// Tag mirrors the structure persisted by SaveTag so we can decode it here
// without a circular dependency.
type Tag struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// loadRawTags reads the raw tag file used internally.
func loadRawTags(dir string) ([]Tag, error) {
	p := tagFilePath(dir)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var tags []Tag
	for _, line := range splitLines(string(data)) {
		var t Tag
		if err := json.Unmarshal([]byte(line), &t); err != nil {
			continue
		}
		tags = append(tags, t)
	}
	return tags, nil
}
