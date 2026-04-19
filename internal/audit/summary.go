package audit

import (
	"fmt"
	"strings"
	"time"
)

// Summary holds aggregated stats from audit log entries.
type Summary struct {
	Total     int
	Succeeded int
	Failed    int
	Paths     []string
	LastSync  time.Time
}

// Summarize aggregates a slice of Entry into a Summary.
func Summarize(entries []Entry) Summary {
	s := Summary{}
	seen := map[string]bool{}

	for _, e := range entries {
		s.Total++
		if strings.EqualFold(e.Status, "success") {
			s.Succeeded++
		} else {
			s.Failed++
		}
		if !seen[e.SecretPath] {
			seen[e.SecretPath] = true
			s.Paths = append(s.Paths, e.SecretPath)
		}
		if e.Timestamp.After(s.LastSync) {
			s.LastSync = e.Timestamp
		}
	}
	return s
}

// String returns a human-readable summary.
func (s Summary) String() string {
	if s.Total == 0 {
		return "No audit entries found."
	}
	return fmt.Sprintf(
		"Total: %d | Succeeded: %d | Failed: %d | Paths: %s | Last sync: %s",
		s.Total, s.Succeeded, s.Failed,
		strings.Join(s.Paths, ", "),
		s.LastSync.Format(time.RFC3339),
	)
}
