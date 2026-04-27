package audit

import (
	"time"
)

// StalenessReport holds staleness information for a single secret path.
type StalenessReport struct {
	Path       string
	LastSynced time.Time
	AgeDays    int
	IsStale    bool
}

// StalenessOptions controls what is considered stale.
type StalenessOptions struct {
	ThresholdDays int // secrets older than this are considered stale
}

// CheckStaleness evaluates each unique path in entries and returns a report
// per path indicating whether it has exceeded the staleness threshold.
func CheckStaleness(entries []Entry, opts StalenessOptions) []StalenessReport {
	if opts.ThresholdDays <= 0 {
		opts.ThresholdDays = 30
	}

	// Find the most recent sync time per path.
	latest := make(map[string]time.Time)
	for _, e := range entries {
		if e.Action != "sync" && e.Action != "write" {
			continue
		}
		if t, ok := latest[e.Path]; !ok || e.Timestamp.After(t) {
			latest[e.Path] = e.Timestamp
		}
	}

	now := time.Now().UTC()
	reports := make([]StalenessReport, 0, len(latest))
	for path, lastSynced := range latest {
		ageDays := int(now.Sub(lastSynced).Hours() / 24)
		reports = append(reports, StalenessReport{
			Path:       path,
			LastSynced: lastSynced,
			AgeDays:    ageDays,
			IsStale:    ageDays >= opts.ThresholdDays,
		})
	}

	// Sort by path for deterministic output.
	for i := 1; i < len(reports); i++ {
		for j := i; j > 0 && reports[j].Path < reports[j-1].Path; j-- {
			reports[j], reports[j-1] = reports[j-1], reports[j]
		}
	}

	return reports
}
