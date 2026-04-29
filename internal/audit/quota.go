package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// QuotaPolicy defines limits for secret sync activity.
type QuotaPolicy struct {
	MaxSyncsPerHour int `json:"max_syncs_per_hour"`
	MaxKeysPerSync  int `json:"max_keys_per_sync"`
	MaxPathsPerDay  int `json:"max_paths_per_day"`
}

// QuotaViolation describes a single quota breach.
type QuotaViolation struct {
	Rule    string
	Current int
	Limit   int
}

func (v QuotaViolation) String() string {
	return fmt.Sprintf("%s: current=%d limit=%d", v.Rule, v.Current, v.Limit)
}

// QuotaResult holds the outcome of a quota check.
type QuotaResult struct {
	Violations []QuotaViolation
}

func (r QuotaResult) OK() bool { return len(r.Violations) == 0 }

func quotaPolicyFilePath(dir string) string {
	return filepath.Join(dir, ".vaultpull_quota.json")
}

// SaveQuotaPolicy persists a QuotaPolicy to disk.
func SaveQuotaPolicy(dir string, p QuotaPolicy) error {
	if p.MaxSyncsPerHour < 0 || p.MaxKeysPerSync < 0 || p.MaxPathsPerDay < 0 {
		return fmt.Errorf("quota values must be non-negative")
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(quotaPolicyFilePath(dir), data, 0600)
}

// LoadQuotaPolicy reads a QuotaPolicy from disk.
func LoadQuotaPolicy(dir string) (QuotaPolicy, error) {
	data, err := os.ReadFile(quotaPolicyFilePath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return QuotaPolicy{}, nil
		}
		return QuotaPolicy{}, err
	}
	var p QuotaPolicy
	return p, json.Unmarshal(data, &p)
}

// CheckQuota evaluates entries against the given policy.
func CheckQuota(entries []Entry, p QuotaPolicy) QuotaResult {
	now := time.Now().UTC()
	hourAgo := now.Add(-time.Hour)
	dayAgo := now.Add(-24 * time.Hour)

	var result QuotaResult

	if p.MaxSyncsPerHour > 0 {
		count := 0
		for _, e := range entries {
			if e.Action == "sync" && e.Timestamp.After(hourAgo) {
				count++
			}
		}
		if count > p.MaxSyncsPerHour {
			result.Violations = append(result.Violations, QuotaViolation{
				Rule: "max_syncs_per_hour", Current: count, Limit: p.MaxSyncsPerHour,
			})
		}
	}

	if p.MaxKeysPerSync > 0 {
		for _, e := range entries {
			if e.Action == "sync" && len(e.Keys) > p.MaxKeysPerSync {
				result.Violations = append(result.Violations, QuotaViolation{
					Rule: "max_keys_per_sync", Current: len(e.Keys), Limit: p.MaxKeysPerSync,
				})
				break
			}
		}
	}

	if p.MaxPathsPerDay > 0 {
		paths := map[string]struct{}{}
		for _, e := range entries {
			if e.Action == "sync" && e.Timestamp.After(dayAgo) {
				paths[e.Path] = struct{}{}
			}
		}
		if len(paths) > p.MaxPathsPerDay {
			result.Violations = append(result.Violations, QuotaViolation{
				Rule: "max_paths_per_day", Current: len(paths), Limit: p.MaxPathsPerDay,
			})
		}
	}

	return result
}
