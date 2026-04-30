package audit

import (
	"math"
	"sort"
	"time"
)

// VelocityResult holds sync frequency metrics for a single secret path.
type VelocityResult struct {
	Path           string
	SyncCount      int
	WindowDays     int
	SyncsPerDay    float64
	PeakDay        string // YYYY-MM-DD of highest activity
	PeakDayCount   int
}

// Velocity computes sync-per-day rates for each path over the given window.
func Velocity(entries []Entry, windowDays int) []VelocityResult {
	if len(entries) == 0 || windowDays <= 0 {
		return nil
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -windowDays)

	type dayBucket map[string]int // date -> count
	pathDays := map[string]dayBucket{}
	pathCount := map[string]int{}

	for _, e := range entries {
		if e.Action != "sync" {
			continue
		}
		t, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil || t.Before(cutoff) {
			continue
		}
		day := t.UTC().Format("2006-01-02")
		if pathDays[e.Path] == nil {
			pathDays[e.Path] = dayBucket{}
		}
		pathDays[e.Path][day]++
		pathCount[e.Path]++
	}

	results := make([]VelocityResult, 0, len(pathDays))
	for path, days := range pathDays {
		peakDay, peakCount := "", 0
		for d, c := range days {
			if c > peakCount || (c == peakCount && d > peakDay) {
				peakDay, peakCount = d, c
			}
		}
		rate := math.Round(float64(pathCount[path])/float64(windowDays)*100) / 100
		results = append(results, VelocityResult{
			Path:         path,
			SyncCount:    pathCount[path],
			WindowDays:   windowDays,
			SyncsPerDay:  rate,
			PeakDay:      peakDay,
			PeakDayCount: peakCount,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].SyncsPerDay != results[j].SyncsPerDay {
			return results[i].SyncsPerDay > results[j].SyncsPerDay
		}
		return results[i].Path < results[j].Path
	})
	return results
}
