package audit

import (
	"strings"
)

// SearchOptions defines criteria for searching audit log entries.
type SearchOptions struct {
	// KeyContains filters entries where the key contains this substring (case-insensitive).
	KeyContains string
	// PathContains filters entries where the vault path contains this substring (case-insensitive).
	PathContains string
	// ActionIn filters entries whose action is in this set. Empty means all actions.
	ActionIn []string
}

// Search returns entries from the audit log that match all provided search criteria.
// An empty SearchOptions returns all entries unchanged.
func Search(logPath string, opts SearchOptions) ([]Entry, error) {
	entries, err := ReadAll(logPath)
	if err != nil {
		return nil, err
	}

	actionSet := make(map[string]struct{}, len(opts.ActionIn))
	for _, a := range opts.ActionIn {
		actionSet[strings.ToLower(a)] = struct{}{}
	}

	var results []Entry
	for _, e := range entries {
		if opts.KeyContains != "" {
			if !strings.Contains(strings.ToLower(e.Key), strings.ToLower(opts.KeyContains)) {
				continue
			}
		}
		if opts.PathContains != "" {
			if !strings.Contains(strings.ToLower(e.Path), strings.ToLower(opts.PathContains)) {
				continue
			}
		}
		if len(actionSet) > 0 {
			if _, ok := actionSet[strings.ToLower(e.Action)]; !ok {
				continue
			}
		}
		results = append(results, e)
	}
	return results, nil
}
