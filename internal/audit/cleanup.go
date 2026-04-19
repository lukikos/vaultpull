package audit

import (
	"os"
	"time"
)

// Cleanup removes audit log entries older than the given retention duration.
// It rewrites the log file in place, preserving only recent entries.
func Cleanup(logPath string, retention time.Duration) (int, error) {
	entries, err := ReadAll(logPath)
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return 0, nil
	}

	cutoff := time.Now().UTC().Add(-retention)
	var kept []Entry
	for _, e := range entries {
		if e.Timestamp.After(cutoff) {
			kept = append(kept, e)
		}
	}

	removed := len(entries) - len(kept)
	if removed == 0 {
		return 0, nil
	}

	// Rewrite file with only kept entries
	if err := os.Remove(logPath); err != nil && !os.IsNotExist(err) {
		return 0, err
	}

	if len(kept) == 0 {
		return removed, nil
	}

	logger, err := NewLogger(logPath)
	if err != nil {
		return 0, err
	}

	for _, e := range kept {
		if err := logger.Record(e.SecretPath, e.KeysWritten, e.OutputFile); err != nil {
			return 0, err
		}
	}

	return removed, nil
}
