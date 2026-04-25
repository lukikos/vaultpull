package audit

import (
	"fmt"
	"sort"
)

// BaselineDiffEntry describes a single key change between a baseline and current state.
type BaselineDiffEntry struct {
	Key    string
	Action string // added, removed, changed, unchanged
}

// BaselineDiff compares a saved baseline against a current secrets map and returns
// the list of differences, sorted by key.
func BaselineDiff(baseline *Baseline, current map[string]string) []BaselineDiffEntry {
	if baseline == nil {
		return nil
	}
	seen := make(map[string]bool)
	var result []BaselineDiffEntry

	for k, oldVal := range baseline.Keys {
		seen[k] = true
		newVal, exists := current[k]
		switch {
		case !exists:
			result = append(result, BaselineDiffEntry{Key: k, Action: "removed"})
		case newVal != oldVal:
			result = append(result, BaselineDiffEntry{Key: k, Action: "changed"})
		default:
			result = append(result, BaselineDiffEntry{Key: k, Action: "unchanged"})
		}
	}

	for k := range current {
		if !seen[k] {
			result = append(result, BaselineDiffEntry{Key: k, Action: "added"})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

// SummarizeBaselineDiff returns a human-readable summary string.
func SummarizeBaselineDiff(entries []BaselineDiffEntry) string {
	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Action]++
	}
	return fmt.Sprintf("added=%d changed=%d removed=%d unchanged=%d",
		counts["added"], counts["changed"], counts["removed"], counts["unchanged"])
}
