package audit

import (
	"fmt"
	"sort"
	"time"
)

// RewindResult holds the reconstructed state of secrets at a given point in time.
type RewindResult struct {
	AsOf  time.Time
	Path  string
	State map[string]string
}

// Rewind reconstructs the state of secrets for a given path as of a specific
// timestamp by replaying audit log entries up to (and including) that time.
func Rewind(logPath, secretPath string, asOf time.Time) (*RewindResult, error) {
	if secretPath == "" {
		return nil, fmt.Errorf("secret path must not be empty")
	}
	if asOf.IsZero() {
		return nil, fmt.Errorf("asOf time must not be zero")
	}

	entries, err := ReadAll(logPath)
	if err != nil {
		return nil, fmt.Errorf("reading audit log: %w", err)
	}

	// Filter entries for the target path at or before asOf, sorted by time.
	var relevant []Entry
	for _, e := range entries {
		if e.Path == secretPath && !e.Timestamp.After(asOf) {
			relevant = append(relevant, e)
		}
	}

	if len(relevant) == 0 {
		return &RewindResult{
			AsOf:  asOf,
			Path:  secretPath,
			State: map[string]string{},
		}, nil
	}

	sort.Slice(relevant, func(i, j int) bool {
		return relevant[i].Timestamp.Before(relevant[j].Timestamp)
	})

	state := make(map[string]string)
	for _, e := range relevant {
		switch e.Action {
		case "sync":
			for k, v := range e.Keys {
				state[k] = v
			}
		case "delete":
			for k := range e.Keys {
				delete(state, k)
			}
		}
	}

	return &RewindResult{
		AsOf:  asOf,
		Path:  secretPath,
		State: state,
	}, nil
}
