package audit

import (
	"time"
)

// ExpiryResult holds the expiry status for a single secret path.
type ExpiryResult struct {
	Path      string
	Key       string
	ExpiresAt time.Time
	Expired   bool
	DaysLeft  int
}

// ExpiryConfig controls thresholds for expiry warnings.
type ExpiryConfig struct {
	WarnWithinDays int // warn if expiry is within this many days
}

// DefaultExpiryConfig returns sensible defaults.
func DefaultExpiryConfig() ExpiryConfig {
	return ExpiryConfig{
		WarnWithinDays: 30,
	}
}

// CheckExpiry inspects audit entries for keys that carry an expiry timestamp
// (stored as metadata in the Value field as "expires=<RFC3339>") and returns
// expiry results for each such key.
func CheckExpiry(entries []Entry, cfg ExpiryConfig) []ExpiryResult {
	if len(entries) == 0 {
		return nil
	}

	now := time.Now().UTC()
	warnBefore := now.Add(time.Duration(cfg.WarnWithinDays) * 24 * time.Hour)

	// track latest entry per path+key
	type pk struct{ path, key string }
	latest := make(map[pk]Entry)
	for _, e := range entries {
		if e.Key == "" {
			continue
		}
		k := pk{e.Path, e.Key}
		if prev, ok := latest[k]; !ok || e.Timestamp.After(prev.Timestamp) {
			latest[k] = e
		}
	}

	var results []ExpiryResult
	for _, e := range latest {
		expiry, ok := parseExpiry(e.Value)
		if !ok {
			continue
		}
		if expiry.After(warnBefore) && expiry.After(now) {
			continue
		}
		daysLeft := int(time.Until(expiry).Hours() / 24)
		results = append(results, ExpiryResult{
			Path:      e.Path,
			Key:       e.Key,
			ExpiresAt: expiry,
			Expired:   expiry.Before(now),
			DaysLeft:  daysLeft,
		})
	}
	return results
}

// parseExpiry extracts an RFC3339 timestamp from a value formatted as
// "expires=<RFC3339>" or any string that is itself a valid RFC3339 time.
func parseExpiry(value string) (time.Time, bool) {
	const prefix = "expires="
	if len(value) > len(prefix) && value[:len(prefix)] == prefix {
		t, err := time.Parse(time.RFC3339, value[len(prefix):])
		if err == nil {
			return t.UTC(), true
		}
	}
	t, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return t.UTC(), true
	}
	return time.Time{}, false
}
