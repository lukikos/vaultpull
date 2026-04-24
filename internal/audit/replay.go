package audit

import (
	"fmt"
	"sort"
	"time"
)

// ReplayResult holds the reconstructed state of secrets at a point in time.
type ReplayResult struct {
	AsOf    time.Time
	Path    string
	Secrets map[string]string
}

// Replay reconstructs the state of secrets for a given vault path as of a
// specific point in time by replaying audit log entries in order.
func Replay(logPath, secretPath string, asOf time.Time) (*ReplayResult, error) {
	entries, err := ReadAll(logPath)
	if err != nil {
		return nil, fmt.Errorf("replay: read log: %w", err)
	}

	// Filter to the target path and entries up to asOf.
	var relevant []Entry
	for _, e := range entries {
		if e.Path != secretPath {
			continue
		}
		if !e.Timestamp.After(asOf) {
			relevant = append(relevant, e)
		}
	}

	if len(relevant) == 0 {
		return &ReplayResult{
			AsOf:    asOf,
			Path:    secretPath,
			Secrets: map[string]string{},
		}, nil
	}

	// Sort ascending by timestamp so we apply changes in order.
	sort.Slice(relevant, func(i, j int) bool {
		return relevant[i].Timestamp.Before(relevant[j].Timestamp)
	})

	state := map[string]string{}
	for _, e := range relevant {
		switch e.Action {
		case "added", "updated":
			state[e.Key] = e.NewValue
		case "removed":
			delete(state, e.Key)
		}
	}

	return &ReplayResult{
		AsOf:    asOf,
		Path:    secretPath,
		Secrets: state,
	}, nil
}
