package audit

import (
	"sort"
	"time"
)

// HeatmapBucket represents sync activity for a single hour-of-day bucket.
type HeatmapBucket struct {
	Hour  int // 0-23
	Count int
}

// HeatmapResult holds the activity distribution across hours for a single path.
type HeatmapResult struct {
	Path    string
	Buckets []HeatmapBucket // always 24 entries, one per hour
	PeakHour int
	PeakCount int
}

// Heatmap computes hourly sync activity distribution per secret path.
// Only entries with action "sync" are considered.
func Heatmap(entries []Entry) []HeatmapResult {
	if len(entries) == 0 {
		return nil
	}

	type hourCounts [24]int
	pathHours := make(map[string]*hourCounts)

	for _, e := range entries {
		if e.Action != "sync" {
			continue
		}
		if _, ok := pathHours[e.Path]; !ok {
			pathHours[e.Path] = &hourCounts{}
		}
		hour := e.Timestamp.UTC().Hour()
		pathHours[e.Path][hour]++
	}

	if len(pathHours) == 0 {
		return nil
	}

	paths := make([]string, 0, len(pathHours))
	for p := range pathHours {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	results := make([]HeatmapResult, 0, len(paths))
	for _, p := range paths {
		counts := pathHours[p]
		buckets := make([]HeatmapBucket, 24)
		peakHour, peakCount := 0, 0
		for h := 0; h < 24; h++ {
			buckets[h] = HeatmapBucket{Hour: h, Count: counts[h]}
			if counts[h] > peakCount {
				peakCount = counts[h]
				peakHour = h
			}
		}
		results = append(results, HeatmapResult{
			Path:      p,
			Buckets:   buckets,
			PeakHour:  peakHour,
			PeakCount: peakCount,
		})
	}
	return results
}

// makeHeatmapEntry is a helper for tests.
func makeHeatmapEntry(path string, hour int) Entry {
	base := time.Date(2024, 1, 15, hour, 0, 0, 0, time.UTC)
	return Entry{Path: path, Action: "sync", Timestamp: base}
}
