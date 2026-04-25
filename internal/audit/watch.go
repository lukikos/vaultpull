package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// WatchEvent represents a change detected during a watch poll cycle.
type WatchEvent struct {
	Path      string    `json:"path"`
	Key       string    `json:"key"`
	Action    string    `json:"action"`
	DetectedAt time.Time `json:"detected_at"`
}

// WatchOptions configures the Watch behaviour.
type WatchOptions struct {
	LogFile  string
	Path     string
	Interval time.Duration
	MaxPolls int // 0 means run forever
}

// Watch polls the audit log for new entries matching the given path and
// emits WatchEvents to the returned channel. The caller must close done
// to stop the watcher.
func Watch(opts WatchOptions, done <-chan struct{}) (<-chan WatchEvent, error) {
	if opts.Interval <= 0 {
		opts.Interval = 5 * time.Second
	}

	events := make(chan WatchEvent, 16)

	go func() {
		defer close(events)

		var lastSeen time.Time
		polls := 0

		for {
			select {
			case <-done:
				return
			case <-time.After(opts.Interval):
			}

			entries, err := readEntriesFrom(opts.LogFile)
			if err != nil {
				// log file may not exist yet; keep polling
				polls++
				if opts.MaxPolls > 0 && polls >= opts.MaxPolls {
					return
				}
				continue
			}

			for _, e := range entries {
				if opts.Path != "" && e.Path != opts.Path {
					continue
				}
				if e.Timestamp.After(lastSeen) {
					events <- WatchEvent{
						Path:       e.Path,
						Key:        e.Key,
						Action:     e.Action,
						DetectedAt: time.Now(),
					}
					lastSeen = e.Timestamp
				}
			}

			polls++
			if opts.MaxPolls > 0 && polls >= opts.MaxPolls {
				return
			}
		}
	}()

	return events, nil
}

// readEntriesFrom reads all log entries from the given file path.
func readEntriesFrom(logFile string) ([]Entry, error) {
	data, err := os.ReadFile(logFile)
	if err != nil {
		return nil, fmt.Errorf("read log: %w", err)
	}
	var entries []Entry
	for _, line := range splitLines(string(data)) {
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}
