package audit

import (
	"sort"
	"time"
)

// PathStats holds aggregated statistics for a single secret path.
type PathStats struct {
	Path       string
	TotalSyncs int
	Added      int
	Updated    int
	Unchanged  int
	FirstSeen  time.Time
	LastSeen   time.Time
}

// Stats returns per-path aggregated statistics from a slice of log entries.
func Stats(entries []Entry) []PathStats {
	type bucket struct {
		total     int
		added     int
		updated   int
		unchanged int
		first     time.Time
		last      time.Time
	}

	m := make(map[string]*bucket)

	for _, e := range entries {
		b, ok := m[e.Path]
		if !ok {
			b = &bucket{first: e.Timestamp, last: e.Timestamp}
			m[e.Path] = b
		}

		b.total++

		switch e.Action {
		case "added":
			b.added++
		case "updated":
			b.updated++
		case "unchanged":
			b.unchanged++
		}

		if e.Timestamp.Before(b.first) {
			b.first = e.Timestamp
		}
		if e.Timestamp.After(b.last) {
			b.last = e.Timestamp
		}
	}

	result := make([]PathStats, 0, len(m))
	for path, b := range m {
		result = append(result, PathStats{
			Path:       path,
			TotalSyncs: b.total,
			Added:      b.added,
			Updated:    b.updated,
			Unchanged:  b.unchanged,
			FirstSeen:  b.first,
			LastSeen:   b.last,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Path < result[j].Path
	})

	return result
}
