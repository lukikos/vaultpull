package audit

import (
	"testing"
	"time"
)

func makeForecastEntry(path, action string, ts time.Time) Entry {
	return Entry{
		Path:      path,
		Action:    action,
		Timestamp: ts,
	}
}

func TestForecast_Empty(t *testing.T) {
	results := Forecast(nil)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestForecast_SingleEntry_NoInterval(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeForecastEntry("secret/app", "sync", now),
	}
	results := Forecast(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].PredictedNext.IsZero() {
		t.Errorf("expected zero PredictedNext for single-sample path")
	}
	if results[0].SampleCount != 1 {
		t.Errorf("expected SampleCount=1, got %d", results[0].SampleCount)
	}
}

func TestForecast_PredictedNext(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []Entry{
		makeForecastEntry("secret/app", "sync", base),
		makeForecastEntry("secret/app", "sync", base.Add(24*time.Hour)),
		makeForecastEntry("secret/app", "sync", base.Add(48*time.Hour)),
	}
	results := Forecast(entries)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.AvgIntervalH != 24.0 {
		t.Errorf("expected AvgIntervalH=24, got %.2f", r.AvgIntervalH)
	}
	expected := base.Add(72 * time.Hour)
	if !r.PredictedNext.Equal(expected) {
		t.Errorf("expected PredictedNext=%v, got %v", expected, r.PredictedNext)
	}
}

func TestForecast_IgnoresNonSyncActions(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []Entry{
		makeForecastEntry("secret/app", "read", base),
		makeForecastEntry("secret/app", "delete", base.Add(12*time.Hour)),
	}
	results := Forecast(entries)
	if len(results) != 0 {
		t.Errorf("expected 0 results when no sync entries, got %d", len(results))
	}
}

func TestForecast_MultiplePathsSorted(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []Entry{
		makeForecastEntry("secret/z", "sync", base),
		makeForecastEntry("secret/a", "sync", base),
		makeForecastEntry("secret/m", "sync", base),
	}
	results := Forecast(entries)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Path != "secret/a" || results[1].Path != "secret/m" || results[2].Path != "secret/z" {
		t.Errorf("results not sorted by path: %v", results)
	}
}
