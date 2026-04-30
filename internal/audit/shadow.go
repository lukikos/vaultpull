package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ShadowEntry records a snapshot of a secret's value hash at a point in time,
// enabling detection of out-of-band changes between vaultpull syncs.
type ShadowEntry struct {
	Path      string    `json:"path"`
	Key       string    `json:"key"`
	ValueHash string    `json:"value_hash"`
	RecordedAt time.Time `json:"recorded_at"`
}

// ShadowReport describes the result of comparing current state to shadow.
type ShadowReport struct {
	Path    string
	Key     string
	Status  string // "match", "changed", "missing"
	Message string
}

func shadowFilePath(dir, name string) string {
	return filepath.Join(dir, fmt.Sprintf(".shadow_%s.json", name))
}

// SaveShadow persists a set of shadow entries under a named shadow file.
func SaveShadow(dir, name string, entries []ShadowEntry) error {
	if name == "" {
		return fmt.Errorf("shadow name must not be empty")
	}
	if len(entries) == 0 {
		return fmt.Errorf("no entries to shadow")
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal shadow: %w", err)
	}
	return os.WriteFile(shadowFilePath(dir, name), data, 0600)
}

// LoadShadow reads a named shadow file and returns its entries.
func LoadShadow(dir, name string) ([]ShadowEntry, error) {
	if name == "" {
		return nil, fmt.Errorf("shadow name must not be empty")
	}
	data, err := os.ReadFile(shadowFilePath(dir, name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("shadow %q not found", name)
		}
		return nil, fmt.Errorf("read shadow: %w", err)
	}
	var entries []ShadowEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse shadow: %w", err)
	}
	return entries, nil
}

// CompareShadow compares a current map of key→valueHash against a saved shadow
// and returns a report of matches, changes, and missing keys.
func CompareShadow(path string, current map[string]string, shadow []ShadowEntry) []ShadowReport {
	var reports []ShadowReport
	for _, s := range shadow {
		if s.Path != path {
			continue
		}
		hash, ok := current[s.Key]
		if !ok {
			reports = append(reports, ShadowReport{
				Path:    path,
				Key:     s.Key,
				Status:  "missing",
				Message: "key no longer present in current state",
			})
			continue
		}
		if hash != s.ValueHash {
			reports = append(reports, ShadowReport{
				Path:    path,
				Key:     s.Key,
				Status:  "changed",
				Message: "value hash differs from shadow",
			})
		} else {
			reports = append(reports, ShadowReport{
				Path:   path,
				Key:    s.Key,
				Status: "match",
			})
		}
	}
	return reports
}
