package audit

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// TTLPolicy defines per-path time-to-live durations for secrets.
type TTLPolicy struct {
	Path    string        `json:"path"`
	TTL     time.Duration `json:"ttl"`
	Created time.Time     `json:"created"`
}

// TTLResult reports whether a path's secrets are still within TTL.
type TTLResult struct {
	Path    string
	TTL     time.Duration
	LastSync time.Time
	Expired bool
	Age     time.Duration
}

func ttlPolicyFilePath(dir string) string {
	return filepath.Join(dir, ".vaultpull_ttl.json")
}

// SaveTTLPolicy persists a TTL policy for a given path.
func SaveTTLPolicy(dir, path string, ttl time.Duration) error {
	if path == "" {
		return errors.New("path must not be empty")
	}
	if ttl <= 0 {
		return errors.New("ttl must be positive")
	}

	policies, _ := LoadTTLPolicies(dir)
	updated := false
	for i, p := range policies {
		if p.Path == path {
			policies[i].TTL = ttl
			updated = true
			break
		}
	}
	if !updated {
		policies = append(policies, TTLPolicy{Path: path, TTL: ttl, Created: time.Now().UTC()})
	}

	data, err := json.MarshalIndent(policies, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ttlPolicyFilePath(dir), data, 0o600)
}

// LoadTTLPolicies reads all saved TTL policies from disk.
func LoadTTLPolicies(dir string) ([]TTLPolicy, error) {
	data, err := os.ReadFile(ttlPolicyFilePath(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var policies []TTLPolicy
	return policies, json.Unmarshal(data, &policies)
}

// CheckTTL evaluates each path's last sync time against its TTL policy.
func CheckTTL(dir string, entries []Entry) ([]TTLResult, error) {
	policies, err := LoadTTLPolicies(dir)
	if err != nil {
		return nil, err
	}

	lastSync := map[string]time.Time{}
	for _, e := range entries {
		if e.Action != "sync" {
			continue
		}
		if t, ok := lastSync[e.Path]; !ok || e.Timestamp.After(t) {
			lastSync[e.Path] = e.Timestamp
		}
	}

	now := time.Now().UTC()
	var results []TTLResult
	for _, p := range policies {
		ls, seen := lastSync[p.Path]
		var age time.Duration
		var expired bool
		if seen {
			age = now.Sub(ls)
			expired = age > p.TTL
		} else {
			expired = true
		}
		results = append(results, TTLResult{
			Path:     p.Path,
			TTL:      p.TTL,
			LastSync: ls,
			Expired:  expired,
			Age:      age,
		})
	}
	return results, nil
}
