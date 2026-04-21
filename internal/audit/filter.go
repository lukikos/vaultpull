package audit

import "time"

// FilterOptions controls which log entries are returned.
type FilterOptions struct {
	Path      string    // filter by secret path (empty = all)
	Action    string    // filter by action (empty = all)
	Since     time.Time // filter entries at or after this time (zero = no lower bound)
	Until     time.Time // filter entries before or at this time (zero = no upper bound)
}

// Filter returns only the log entries that match all non-zero criteria in opts.
func Filter(entries []LogEntry, opts FilterOptions) []LogEntry {
	result := make([]LogEntry, 0, len(entries))
	for _, e := range entries {
		if opts.Path != "" && e.Path != opts.Path {
			continue
		}
		if opts.Action != "" && e.Action != opts.Action {
			continue
		}
		if !opts.Since.IsZero() && e.Timestamp.Before(opts.Since) {
			continue
		}
		if !opts.Until.IsZero() && e.Timestamp.After(opts.Until) {
			continue
		}
		result = append(result, e)
	}
	return result
}
