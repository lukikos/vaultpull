package audit

import (
	"sort"
	"time"
)

// DriftResult describes the drift state for a single secret path.
type DriftResult struct {
	Path        string
	LastSync    time.Time
	AgeHours    float64
	SyncCount   int
	HasDrift    bool
	Reason      string
}

// DriftConfig controls thresholds for drift detection.
type DriftConfig struct {
	// MaxAgeHours is the maximum acceptable time since last sync before drift is flagged.
	MaxAgeHours float64
	// MinSyncCount is the minimum number of syncs expected before flagging low activity.
	MinSyncCount int
}

// DefaultDriftConfig returns sensible defaults.
func DefaultDriftConfig() DriftConfig {
	return DriftConfig{
		MaxAgeHours:  48,
		MinSyncCount: 1,
	}
}

// DetectDrift analyses audit entries and returns a DriftResult per path.
func DetectDrift(entries []Entry, cfg DriftConfig, now time.Time) []DriftResult {
	type pathState struct {
		lastSync  time.Time
		syncs     int
	}

	states := map[string]*pathState{}
	for _, e := range entries {
		if e.Action != "sync" {
			continue
		}
		ps, ok := states[e.Path]
		if !ok {
			ps = &pathState{}
			states[e.Path] = ps
		}
		ps.syncs++
		if e.Timestamp.After(ps.lastSync) {
			ps.lastSync = e.Timestamp
		}
	}

	results := make([]DriftResult, 0, len(states))
	for path, ps := range states {
		ageHours := now.Sub(ps.lastSync).Hours()
		hasDrift := false
		reason := ""

		if ageHours > cfg.MaxAgeHours {
			hasDrift = true
			reason = "last sync exceeded max age threshold"
		} else if ps.syncs < cfg.MinSyncCount {
			hasDrift = true
			reason = "sync count below minimum threshold"
		}

		results = append(results, DriftResult{
			Path:      path,
			LastSync:  ps.lastSync,
			AgeHours:  ageHours,
			SyncCount: ps.syncs,
			HasDrift:  hasDrift,
			Reason:    reason,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})
	return results
}
