package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// RetentionPolicy defines rules for how long audit entries should be kept.
type RetentionPolicy struct {
	MaxAgeDays int `json:"max_age_days"`
	MaxEntries int `json:"max_entries"`
}

// RetentionResult summarizes the outcome of a retention enforcement run.
type RetentionResult struct {
	Removed  int
	Retained int
}

func retentionFilePath(dir string) string {
	return dir + "/retention.json"
}

// SaveRetentionPolicy persists a RetentionPolicy to disk.
func SaveRetentionPolicy(dir string, policy RetentionPolicy) error {
	if policy.MaxAgeDays < 0 || policy.MaxEntries < 0 {
		return fmt.Errorf("retention policy values must be non-negative")
	}
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal retention policy: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create retention dir: %w", err)
	}
	return os.WriteFile(retentionFilePath(dir), data, 0o600)
}

// LoadRetentionPolicy loads a previously saved RetentionPolicy.
func LoadRetentionPolicy(dir string) (RetentionPolicy, error) {
	data, err := os.ReadFile(retentionFilePath(dir))
	if os.IsNotExist(err) {
		return RetentionPolicy{}, fmt.Errorf("no retention policy found")
	}
	if err != nil {
		return RetentionPolicy{}, fmt.Errorf("read retention policy: %w", err)
	}
	var p RetentionPolicy
	if err := json.Unmarshal(data, &p); err != nil {
		return RetentionPolicy{}, fmt.Errorf("parse retention policy: %w", err)
	}
	return p, nil
}

// EnforceRetention applies the given policy to the audit log at logPath,
// removing entries that exceed the age or count limits.
func EnforceRetention(logPath string, policy RetentionPolicy) (RetentionResult, error) {
	entries, err := ReadAll(logPath)
	if os.IsNotExist(err) {
		return RetentionResult{}, nil
	}
	if err != nil {
		return RetentionResult{}, fmt.Errorf("read log: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -policy.MaxAgeDays)
	var kept []Entry
	for _, e := range entries {
		if policy.MaxAgeDays > 0 && e.Timestamp.Before(cutoff) {
			continue
		}
		kept = append(kept, e)
	}

	if policy.MaxEntries > 0 && len(kept) > policy.MaxEntries {
		kept = kept[len(kept)-policy.MaxEntries:]
	}

	removed := len(entries) - len(kept)
	if removed == 0 {
		return RetentionResult{Retained: len(kept)}, nil
	}

	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return RetentionResult{}, fmt.Errorf("open log for write: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, e := range kept {
		if err := enc.Encode(e); err != nil {
			return RetentionResult{}, fmt.Errorf("write entry: %w", err)
		}
	}
	return RetentionResult{Removed: removed, Retained: len(kept)}, nil
}
