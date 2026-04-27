package audit

import (
	"sort"
	"time"
)

// ForecastEntry represents a predicted next-sync time for a secret path.
type ForecastEntry struct {
	Path          string
	AvgIntervalH  float64
	LastSynced    time.Time
	PredictedNext time.Time
	SampleCount   int
}

// Forecast analyzes audit log entries and predicts the next expected sync
// time for each path based on average historical sync intervals.
func Forecast(entries []Entry) []ForecastEntry {
	type pathData struct {
		times []time.Time
	}

	byPath := make(map[string]*pathData)

	for _, e := range entries {
		if e.Action != "sync" {
			continue
		}
		if _, ok := byPath[e.Path]; !ok {
			byPath[e.Path] = &pathData{}
		}
		byPath[e.Path].times = append(byPath[e.Path].times, e.Timestamp)
	}

	var results []ForecastEntry

	for path, data := range byPath {
		times := data.times
		sort.Slice(times, func(i, j int) bool {
			return times[i].Before(times[j])
		})

		if len(times) < 2 {
			// Not enough data to forecast
			results = append(results, ForecastEntry{
				Path:          path,
				AvgIntervalH:  0,
				LastSynced:    times[len(times)-1],
				PredictedNext: time.Time{},
				SampleCount:   len(times),
			})
			continue
		}

		var totalDuration time.Duration
		for i := 1; i < len(times); i++ {
			totalDuration += times[i].Sub(times[i-1])
		}
		avgInterval := totalDuration / time.Duration(len(times)-1)
		avgIntervalH := avgInterval.Hours()

		last := times[len(times)-1]
		predicted := last.Add(avgInterval)

		results = append(results, ForecastEntry{
			Path:          path,
			AvgIntervalH:  avgIntervalH,
			LastSynced:    last,
			PredictedNext: predicted,
			SampleCount:   len(times),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})

	return results
}
