package audit

import (
	"fmt"
	"time"
)

// HealthStatus represents the overall health of a secret path.
type HealthStatus struct {
	Path        string
	Status      string // "healthy", "stale", "error", "unknown"
	LastSync    time.Time
	ErrorCount  int
	TotalSyncs  int
	Message     string
}

// HealthReport holds health statuses for all observed paths.
type HealthReport struct {
	GeneratedAt time.Time
	Statuses    []HealthStatus
	Healthy     int
	Stale       int
	Error       int
	Unknown     int
}

// CheckHealth evaluates the health of all secret paths based on audit entries.
// A path is considered:
//   - "healthy" if synced within staleDays and no recent errors
//   - "stale"   if last sync exceeds staleDays
//   - "error"   if the most recent entry has action "error"
//   - "unknown" if no entries exist for the path
func CheckHealth(entries []Entry, staleDays int) HealthReport {
	report := HealthReport{GeneratedAt: time.Now()}
	if len(entries) == 0 {
		return report
	}

	type pathData struct {
		latest    time.Time
		latestAct string
		errors    int
		total     int
	}

	paths := make(map[string]*pathData)
	for _, e := range entries {
		pd, ok := paths[e.Path]
		if !ok {
			pd = &pathData{}
			paths[e.Path] = pd
		}
		pd.total++
		if e.Action == "error" {
			pd.errors++
		}
		if e.Timestamp.After(pd.latest) {
			pd.latest = e.Timestamp
			pd.latestAct = e.Action
		}
	}

	cutoff := time.Now().AddDate(0, 0, -staleDays)

	for path, pd := range paths {
		hs := HealthStatus{
			Path:       path,
			LastSync:   pd.latest,
			ErrorCount: pd.errors,
			TotalSyncs: pd.total,
		}
		switch {
		case pd.latestAct == "error":
			hs.Status = "error"
			hs.Message = fmt.Sprintf("last operation failed (%d total errors)", pd.errors)
			report.Error++
		case pd.latest.Before(cutoff):
			hs.Status = "stale"
			hs.Message = fmt.Sprintf("last sync was %d days ago", int(time.Since(pd.latest).Hours()/24))
			report.Stale++
		default:
			hs.Status = "healthy"
			hs.Message = "synced recently"
			report.Healthy++
		}
		report.Statuses = append(report.Statuses, hs)
	}

	return report
}
