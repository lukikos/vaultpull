package audit

import (
	"sort"
	"time"
)

// LineageEntry represents a single point in a secret key's history.
type LineageEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Key       string    `json:"key"`
	Action    string    `json:"action"`
	Value     string    `json:"value"`
}

// LineageResult holds the full ordered history for a single key within a path.
type LineageResult struct {
	Path    string         `json:"path"`
	Key     string         `json:"key"`
	History []LineageEntry `json:"history"`
}

// Lineage returns the chronological history of a specific key at a given path.
// If key is empty, all keys at the path are included.
func Lineage(entries []Entry, path, key string) []LineageResult {
	if len(entries) == 0 {
		return nil
	}

	// Group by key.
	type bucket struct {
		path string
		key  string
	}
	grouped := make(map[bucket][]LineageEntry)

	for _, e := range entries {
		if e.Path != path {
			continue
		}
		if key != "" && e.Key != key {
			continue
		}
		if e.Action != "sync" && e.Action != "update" && e.Action != "add" && e.Action != "delete" {
			continue
		}
		b := bucket{path: e.Path, key: e.Key}
		grouped[b] = append(grouped[b], LineageEntry{
			Timestamp: e.Timestamp,
			Path:      e.Path,
			Key:       e.Key,
			Action:    e.Action,
			Value:     e.Value,
		})
	}

	results := make([]LineageResult, 0, len(grouped))
	for b, history := range grouped {
		sort.Slice(history, func(i, j int) bool {
			return history[i].Timestamp.Before(history[j].Timestamp)
		})
		results = append(results, LineageResult{
			Path:    b.path,
			Key:     b.key,
			History: history,
		})
	}

	// Sort results by key name for deterministic output.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Key < results[j].Key
	})

	return results
}
