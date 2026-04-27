package audit

import (
	"math"
	"time"
)

// AnomalyResult holds the result of an anomaly check for a single path.
type AnomalyResult struct {
	Path        string
	MeanInterval float64 // seconds between syncs
	StdDev      float64
	LastSeen    time.Time
	IsAnomaly   bool
	Reason      string
}

// DetectAnomalies identifies paths whose sync frequency deviates significantly
// from their historical pattern (more than 2 standard deviations from the mean).
func DetectAnomalies(entries []LogEntry) []AnomalyResult {
	byPath := map[string][]time.Time{}
	for _, e := range entries {
		if e.Action != "sync" {
			continue
		}
		byPath[e.Path] = append(byPath[e.Path], e.Timestamp)
	}

	var results []AnomalyResult
	for path, times := range byPath {
		if len(times) < 3 {
			continue
		}
		sortTimes(times)
		intervals := computeIntervals(times)
		mean := meanOf(intervals)
		stddev := stddevOf(intervals, mean)

		last := times[len(times)-1]
		result := AnomalyResult{
			Path:        path,
			MeanInterval: mean,
			StdDev:      stddev,
			LastSeen:    last,
		}

		sinceLastSync := time.Since(last).Seconds()
		if stddev > 0 && sinceLastSync > mean+2*stddev {
			result.IsAnomaly = true
			result.Reason = "overdue: last sync significantly later than expected"
		}
		results = append(results, result)
	}
	return results
}

func sortTimes(ts []time.Time) {
	for i := 1; i < len(ts); i++ {
		for j := i; j > 0 && ts[j].Before(ts[j-1]); j-- {
			ts[j], ts[j-1] = ts[j-1], ts[j]
		}
	}
}

func computeIntervals(ts []time.Time) []float64 {
	intervals := make([]float64, len(ts)-1)
	for i := 1; i < len(ts); i++ {
		intervals[i-1] = ts[i].Sub(ts[i-1]).Seconds()
	}
	return intervals
}

func meanOf(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func stddevOf(vals []float64, mean float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		diff := v - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(vals)))
}
