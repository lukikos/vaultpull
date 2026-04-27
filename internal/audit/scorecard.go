package audit

import (
	"fmt"
	"sort"
	"time"
)

// ScorecardResult holds the health score and breakdown for a secret path.
type ScorecardResult struct {
	Path          string
	Score         int // 0–100
	TotalSyncs    int
	RecentSyncs   int // within last 7 days
	LastSyncedAt  time.Time
	StaleDays     int // days since last sync (0 if never stale)
	HasErrors     bool
	Notes         []string
}

// Scorecard evaluates the health of each secret path based on audit log entries.
// A higher score indicates a healthy, frequently-synced path.
func Scorecard(entries []Entry) []ScorecardResult {
	type pathData struct {
		total      int
		recent     int
		lastSynced time.Time
		hasErrors  bool
	}

	now := time.Now()
	cutoff := now.AddDate(0, 0, -7)

	pathMap := make(map[string]*pathData)
	for _, e := range entries {
		pd := pathMap[e.Path]
		if pd == nil {
			pd = &pathData{}
			pathMap[e.Path] = pd
		}
		pd.total++
		if e.Timestamp.After(cutoff) {
			pd.recent++
		}
		if e.Timestamp.After(pd.lastSynced) {
			pd.lastSynced = e.Timestamp
		}
		if e.Action == "error" {
			pd.hasErrors = true
		}
	}

	results := make([]ScorecardResult, 0, len(pathMap))
	for path, pd := range pathMap {
		score := 100
		var notes []string

		staleDays := int(now.Sub(pd.lastSynced).Hours() / 24)
		if staleDays > 30 {
			score -= 40
			notes = append(notes, fmt.Sprintf("not synced in %d days", staleDays))
		} else if staleDays > 7 {
			score -= 20
			notes = append(notes, fmt.Sprintf("not synced in %d days", staleDays))
		}

		if pd.recent == 0 {
			score -= 20
			notes = append(notes, "no syncs in last 7 days")
		}

		if pd.hasErrors {
			score -= 20
			notes = append(notes, "has error entries")
		}

		if score < 0 {
			score = 0
		}

		results = append(results, ScorecardResult{
			Path:         path,
			Score:        score,
			TotalSyncs:   pd.total,
			RecentSyncs:  pd.recent,
			LastSyncedAt: pd.lastSynced,
			StaleDays:    staleDays,
			HasErrors:    pd.hasErrors,
			Notes:        notes,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score < results[j].Score
		}
		return results[i].Path < results[j].Path
	})

	return results
}
