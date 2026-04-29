package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// DigestEntry holds a per-path digest computed from the latest secret values.
type DigestEntry struct {
	Path      string    `json:"path"`
	Digest    string    `json:"digest"`
	KeyCount  int       `json:"key_count"`
	ComputedAt time.Time `json:"computed_at"`
}

// DigestReport is the full output of ComputeDigests.
type DigestReport struct {
	Entries []DigestEntry `json:"entries"`
}

func digestFilePath(dir string) string {
	return filepath.Join(dir, ".vaultpull_digests.json")
}

// ComputeDigests replays the audit log to reconstruct the latest state per
// path and produces a deterministic SHA-256 digest of each path's key/value set.
func ComputeDigests(dir string) (DigestReport, error) {
	entries, err := ReadAll(dir)
	if err != nil {
		return DigestReport{}, fmt.Errorf("digest: read log: %w", err)
	}

	// Group latest state per path using Replay logic.
	paths := uniquePaths(entries)
	var report DigestReport

	for _, path := range paths {
		state, err := Replay(dir, path, time.Now())
		if err != nil {
			return DigestReport{}, fmt.Errorf("digest: replay %s: %w", path, err)
		}
		if len(state) == 0 {
			continue
		}
		d, err := hashState(state)
		if err != nil {
			return DigestReport{}, fmt.Errorf("digest: hash %s: %w", path, err)
		}
		report.Entries = append(report.Entries, DigestEntry{
			Path:       path,
			Digest:     d,
			KeyCount:   len(state),
			ComputedAt: time.Now().UTC(),
		})
	}

	sort.Slice(report.Entries, func(i, j int) bool {
		return report.Entries[i].Path < report.Entries[j].Path
	})
	return report, nil
}

// SaveDigests writes a DigestReport to disk.
func SaveDigests(dir string, report DigestReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("digest: marshal: %w", err)
	}
	return os.WriteFile(digestFilePath(dir), data, 0o600)
}

// LoadDigests reads a previously saved DigestReport from disk.
func LoadDigests(dir string) (DigestReport, error) {
	data, err := os.ReadFile(digestFilePath(dir))
	if os.IsNotExist(err) {
		return DigestReport{}, nil
	}
	if err != nil {
		return DigestReport{}, fmt.Errorf("digest: read file: %w", err)
	}
	var report DigestReport
	if err := json.Unmarshal(data, &report); err != nil {
		return DigestReport{}, fmt.Errorf("digest: unmarshal: %w", err)
	}
	return report, nil
}

func hashState(state map[string]string) (string, error) {
	keys := make([]string, 0, len(state))
	for k := range state {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		if _, err := fmt.Fprintf(h, "%s=%s\n", k, state[k]); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func uniquePaths(entries []LogEntry) []string {
	seen := map[string]struct{}{}
	for _, e := range entries {
		seen[e.Path] = struct{}{}
	}
	paths := make([]string, 0, len(seen))
	for p := range seen {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}
