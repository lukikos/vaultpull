package audit

import (
	"encoding/json"
	"os"
	"sort"
)

// PruneOptions controls which entries are kept during pruning.
type PruneOptions struct {
	// KeepTopN retains only the N most recent entries per secret path.
	KeepTopN int
}

// Prune removes duplicate entries per secret path, keeping only the most
// recent N entries for each path. Returns the number of entries removed.
func Prune(logPath string, opts PruneOptions) (int, error) {
	data, err := os.ReadFile(logPath)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var entries []Entry
	for _, line := range splitLines(data) {
		var e Entry
		if err := json.Unmarshal(line, &e); err == nil {
			entries = append(entries, e)
		}
	}

	originalCount := len(entries)

	// Group entries by secret path.
	byPath := make(map[string][]Entry)
	for _, e := range entries {
		byPath[e.SecretPath] = append(byPath[e.SecretPath], e)
	}

	n := opts.KeepTopN
	if n <= 0 {
		n = 10
	}

	var pruned []Entry
	for _, group := range byPath {
		sort.Slice(group, func(i, j int) bool {
			return group[i].Timestamp.After(group[j].Timestamp)
		})
		if len(group) > n {
			group = group[:n]
		}
		pruned = append(pruned, group...)
	}

	// Re-sort all kept entries chronologically.
	sort.Slice(pruned, func(i, j int) bool {
		return pruned[i].Timestamp.Before(pruned[j].Timestamp)
	})

	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, e := range pruned {
		if err := enc.Encode(e); err != nil {
			return 0, err
		}
	}

	return originalCount - len(pruned), nil
}

// splitLines returns non-empty lines from raw log data.
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			line := data[start:i]
			if len(line) > 0 {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(data) {
		if line := data[start:]; len(line) > 0 {
			lines = append(lines, line)
		}
	}
	return lines
}
