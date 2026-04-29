package audit

import (
	"fmt"
	"time"
)

// BadgeStatus represents the health level for a badge.
type BadgeStatus string

const (
	BadgeStatusOK      BadgeStatus = "ok"
	BadgeStatusWarning BadgeStatus = "warning"
	BadgeStatusError   BadgeStatus = "error"
	BadgeStatusUnknown BadgeStatus = "unknown"
)

// Badge holds the computed badge data for a vault path.
type Badge struct {
	Path        string      `json:"path"`
	Status      BadgeStatus `json:"status"`
	Label       string      `json:"label"`
	Message     string      `json:"message"`
	LastSync    *time.Time  `json:"last_sync,omitempty"`
	GeneratedAt time.Time   `json:"generated_at"`
}

// BadgeConfig controls thresholds used when computing badges.
type BadgeConfig struct {
	StalenessThreshold time.Duration
	ErrorRateThreshold float64 // 0.0–1.0
}

// DefaultBadgeConfig returns sensible defaults.
func DefaultBadgeConfig() BadgeConfig {
	return BadgeConfig{
		StalenessThreshold: 24 * time.Hour,
		ErrorRateThreshold: 0.2,
	}
}

// GenerateBadges computes a Badge for every unique path found in entries.
func GenerateBadges(entries []Entry, cfg BadgeConfig) []Badge {
	if len(entries) == 0 {
		return []Badge{}
	}

	type pathStats struct {
		total  int
		errors int
		last   time.Time
	}

	paths := map[string]*pathStats{}
	for _, e := range entries {
		ps, ok := paths[e.Path]
		if !ok {
			ps = &pathStats{}
			paths[e.Path] = ps
		}
		ps.total++
		if e.Action == "error" {
			ps.errors++
		}
		if e.Timestamp.After(ps.last) {
			ps.last = e.Timestamp
		}
	}

	now := time.Now()
	badges := make([]Badge, 0, len(paths))
	for path, ps := range paths {
		b := Badge{
			Path:        path,
			GeneratedAt: now,
		}
		if !ps.last.IsZero() {
			t := ps.last
			b.LastSync = &t
		}

		errorRate := float64(ps.errors) / float64(ps.total)
		switch {
		case errorRate >= cfg.ErrorRateThreshold:
			b.Status = BadgeStatusError
			b.Label = "sync"
			b.Message = fmt.Sprintf("%.0f%% errors", errorRate*100)
		case ps.last.IsZero() || now.Sub(ps.last) > cfg.StalenessThreshold:
			b.Status = BadgeStatusWarning
			b.Label = "sync"
			b.Message = "stale"
		default:
			b.Status = BadgeStatusOK
			b.Label = "sync"
			b.Message = "ok"
		}
		badges = append(badges, b)
	}
	return badges
}
