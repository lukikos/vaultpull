package audit

import "sort"

// KeyChange describes what happened to a single secret key during a sync.
type KeyChange struct {
	Key    string `json:"key"`
	Action string `json:"action"` // "added", "updated", "unchanged"
}

// Diff compares the previous set of keys with the incoming secrets and
// returns a slice of KeyChange values describing every key.
func Diff(previous map[string]string, incoming map[string]string) []KeyChange {
	changes := make([]KeyChange, 0, len(incoming))

	for k, newVal := range incoming {
		if oldVal, exists := previous[k]; !exists {
			changes = append(changes, KeyChange{Key: k, Action: "added"})
		} else if oldVal != newVal {
			changes = append(changes, KeyChange{Key: k, Action: "updated"})
		} else {
			changes = append(changes, KeyChange{Key: k, Action: "unchanged"})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Key < changes[j].Key
	})

	return changes
}

// DiffSummary holds aggregate counts from a Diff result.
type DiffSummary struct {
	Added     int
	Updated   int
	Unchanged int
}

// Summarize aggregates a slice of KeyChange into a DiffSummary.
func SummarizeDiff(changes []KeyChange) DiffSummary {
	var s DiffSummary
	for _, c := range changes {
		switch c.Action {
		case "added":
			s.Added++
		case "updated":
			s.Updated++
		case "unchanged":
			s.Unchanged++
		}
	}
	return s
}
