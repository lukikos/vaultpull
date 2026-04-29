package audit

import (
	"sort"
	"time"
)

// ThrottleResult holds the analysis for a single secret path.
type ThrottleResult struct {
	Path        string
	SyncCount   int
	WindowStart time.Time
	WindowEnd   time.Time
	Exceeds     bool
	Message     string
}

// ThrottleConfig controls how sync frequency is evaluated.
type ThrottleConfig struct {
	// MaxSyncs is the maximum number of syncs allowed within Window.
	MaxSyncs int
	// Window is the duration of the rolling time window.
	Window time.Duration
}

// DefaultThrottleConfig returns a sensible default throttle configuration.
func DefaultThrottleConfig() ThrottleConfig {
	return ThrottleConfig{
		MaxSyncs: 10,
		Window:   1 * time.Hour,
	}
}

// CheckThrottle analyses audit entries and reports paths that have exceeded
// the allowed sync frequency within the rolling window ending at `now`.
func CheckThrottle(entries []Entry, cfg ThrottleConfig, now time.Time) []ThrottleResult {
	if len(entries) == 0 || cfg.MaxSyncs <= 0 || cfg.Window <= 0 {
		return nil
	}

	windowStart := now.Add(-cfg.Window)

	// Group sync timestamps by path.
	byPath := make(map[string][]time.Time)
	for _, e := range entries {
		if e.Action != "sync" {
			continue
		}
		if e.Timestamp.Before(windowStart) {
			continue
		}
		byPath[e.Path] = append(byPath[e.Path], e.Timestamp)
	}

	paths := make([]string, 0, len(byPath))
	for p := range byPath {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var results []ThrottleResult
	for _, path := range paths {
		times := byPath[path]
		count := len(times)
		exceeds := count > cfg.MaxSyncs
		msg := ""
		if exceeds {
			msg = "sync frequency exceeded: " + path
		}
		results = append(results, ThrottleResult{
			Path:        path,
			SyncCount:   count,
			WindowStart: windowStart,
			WindowEnd:   now,
			Exceeds:     exceeds,
			Message:     msg,
		})
	}
	return results
}
