package audit

import (
	"fmt"
	"time"
)

// AlertLevel represents the severity of an alert.
type AlertLevel string

const (
	AlertInfo    AlertLevel = "info"
	AlertWarning AlertLevel = "warning"
	AlertCritical AlertLevel = "critical"
)

// Alert represents a single alert generated from audit log analysis.
type Alert struct {
	Path      string     `json:"path"`
	Level     AlertLevel `json:"level"`
	Message   string     `json:"message"`
	Triggered time.Time  `json:"triggered"`
}

// AlertConfig defines thresholds used to generate alerts.
type AlertConfig struct {
	// MaxErrorRate is the fraction of entries that may be errors before a critical alert (0.0–1.0).
	MaxErrorRate float64
	// StaleAfter is the duration after which a path with no sync is considered stale.
	StaleAfter time.Duration
	// MaxSyncInterval is the duration beyond which a warning is raised if no sync has occurred.
	MaxSyncInterval time.Duration
}

// DefaultAlertConfig returns sensible defaults.
func DefaultAlertConfig() AlertConfig {
	return AlertConfig{
		MaxErrorRate:    0.1,
		StaleAfter:      48 * time.Hour,
		MaxSyncInterval: 24 * time.Hour,
	}
}

// CheckAlerts analyses audit log entries and returns alerts based on cfg.
func CheckAlerts(entries []Entry, cfg AlertConfig) []Alert {
	if len(entries) == 0 {
		return nil
	}

	type pathStats struct {
		total    int
		errors   int
		lastSync time.Time
	}

	byPath := make(map[string]*pathStats)
	for _, e := range entries {
		ps := byPath[e.Path]
		if ps == nil {
			ps = &pathStats{}
			byPath[e.Path] = ps
		}
		ps.total++
		if e.Action == "error" {
			ps.errors++
		}
		if e.Action == "sync" && e.Timestamp.After(ps.lastSync) {
			ps.lastSync = e.Timestamp
		}
	}

	now := time.Now()
	var alerts []Alert

	for path, ps := range byPath {
		if ps.total > 0 {
			errorRate := float64(ps.errors) / float64(ps.total)
			if errorRate > cfg.MaxErrorRate {
				alerts = append(alerts, Alert{
					Path:      path,
					Level:     AlertCritical,
					Message:   fmt.Sprintf("error rate %.0f%% exceeds threshold %.0f%%", errorRate*100, cfg.MaxErrorRate*100),
					Triggered: now,
				})
			}
		}

		if !ps.lastSync.IsZero() {
			age := now.Sub(ps.lastSync)
			if age > cfg.StaleAfter {
				alerts = append(alerts, Alert{
					Path:      path,
					Level:     AlertCritical,
					Message:   fmt.Sprintf("no sync in %.0fh (stale threshold %.0fh)", age.Hours(), cfg.StaleAfter.Hours()),
					Triggered: now,
				})
			} else if age > cfg.MaxSyncInterval {
				alerts = append(alerts, Alert{
					Path:      path,
					Level:     AlertWarning,
					Message:   fmt.Sprintf("no sync in %.0fh (interval threshold %.0fh)", age.Hours(), cfg.MaxSyncInterval.Hours()),
					Triggered: now,
				})
			}
		}
	}

	return alerts
}
