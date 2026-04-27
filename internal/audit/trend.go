package audit

import (
	"math"
	"sort"
	"time"
)

// TrendPoint represents the sync frequency for a given day.
type TrendPoint struct {
	Date  time.Time
	Count int
}

// TrendResult holds trend analysis for a single secret path.
type TrendResult struct {
	Path          string
	Points        []TrendPoint
	AvgPerDay     float64
	PeakDate      time.Time
	PeakCount     int
	TrendSlope    float64 // positive = increasing, negative = decreasing
}

// Trend analyses sync frequency over time per secret path.
func Trend(entries []Entry, since time.Time) []TrendResult {
	type dayKey struct {
		path string
		day  string
	}

	counts := make(map[dayKey]int)
	paths := make(map[string]struct{})

	for _, e := range entries {
		if e.Action != "sync" {
			continue
		}
		if e.Timestamp.Before(since) {
			continue
		}
		day := e.Timestamp.UTC().Format("2006-01-02")
		counts[dayKey{path: e.Path, day: day}]++
		paths[e.Path] = struct{}{}
	}

	var results []TrendResult
	for path := range paths {
		pointMap := make(map[string]int)
		for k, v := range counts {
			if k.path == path {
				pointMap[k.day] += v
			}
		}

		var points []TrendPoint
		for day, cnt := range pointMap {
			t, _ := time.Parse("2006-01-02", day)
			points = append(points, TrendPoint{Date: t, Count: cnt})
		}
		sort.Slice(points, func(i, j int) bool {
			return points[i].Date.Before(points[j].Date)
		})

		var total, peak int
		var peakDate time.Time
		for _, p := range points {
			total += p.Count
			if p.Count > peak {
				peak = p.Count
				peakDate = p.Date
			}
		}

		avg := 0.0
		if len(points) > 0 {
			avg = float64(total) / float64(len(points))
		}

		results = append(results, TrendResult{
			Path:       path,
			Points:     points,
			AvgPerDay:  math.Round(avg*100) / 100,
			PeakDate:   peakDate,
			PeakCount:  peak,
			TrendSlope: computeSlope(points),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})
	return results
}

// computeSlope returns a simple linear regression slope over the point counts.
func computeSlope(points []TrendPoint) float64 {
	n := float64(len(points))
	if n < 2 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for i, p := range points {
		x := float64(i)
		y := float64(p.Count)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return math.Round(((n*sumXY-sumX*sumY)/denom)*1000) / 1000
}
