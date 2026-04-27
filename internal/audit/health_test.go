package audit

import (
	"testing"
	"time"
)

func makeHealthEntry(path, action string, daysAgo int) Entry {
	return Entry{
		Path:      path,
		Action:    action,
		Timestamp: time.Now().AddDate(0, 0, -daysAgo),
	}
}

func TestCheckHealth_Empty(t *testing.T) {
	report := CheckHealth(nil, 7)
	if len(report.Statuses) != 0 {
		t.Fatalf("expected no statuses, got %d", len(report.Statuses))
	}
	if report.Healthy != 0 || report.Stale != 0 || report.Error != 0 {
		t.Fatal("expected all counts to be zero")
	}
}

func TestCheckHealth_HealthyPath(t *testing.T) {
	entries := []Entry{
		makeHealthEntry("secret/app", "sync", 1),
		makeHealthEntry("secret/app", "sync", 3),
	}
	report := CheckHealth(entries, 7)
	if report.Healthy != 1 {
		t.Fatalf("expected 1 healthy path, got %d", report.Healthy)
	}
	if report.Statuses[0].Status != "healthy" {
		t.Errorf("expected healthy, got %s", report.Statuses[0].Status)
	}
}

func TestCheckHealth_StalePath(t *testing.T) {
	entries := []Entry{
		makeHealthEntry("secret/old", "sync", 15),
	}
	report := CheckHealth(entries, 7)
	if report.Stale != 1 {
		t.Fatalf("expected 1 stale path, got %d", report.Stale)
	}
	if report.Statuses[0].Status != "stale" {
		t.Errorf("expected stale, got %s", report.Statuses[0].Status)
	}
}

func TestCheckHealth_ErrorPath(t *testing.T) {
	entries := []Entry{
		makeHealthEntry("secret/broken", "sync", 1),
		makeHealthEntry("secret/broken", "error", 0),
	}
	report := CheckHealth(entries, 7)
	if report.Error != 1 {
		t.Fatalf("expected 1 error path, got %d", report.Error)
	}
	if report.Statuses[0].Status != "error" {
		t.Errorf("expected error, got %s", report.Statuses[0].Status)
	}
	if report.Statuses[0].ErrorCount != 1 {
		t.Errorf("expected error count 1, got %d", report.Statuses[0].ErrorCount)
	}
}

func TestCheckHealth_MultiplePaths(t *testing.T) {
	entries := []Entry{
		makeHealthEntry("secret/a", "sync", 1),
		makeHealthEntry("secret/b", "sync", 20),
		makeHealthEntry("secret/c", "error", 0),
	}
	report := CheckHealth(entries, 7)
	if report.Healthy+report.Stale+report.Error != 3 {
		t.Fatalf("expected 3 total paths, got %d", report.Healthy+report.Stale+report.Error)
	}
	if report.Healthy != 1 {
		t.Errorf("expected 1 healthy, got %d", report.Healthy)
	}
	if report.Stale != 1 {
		t.Errorf("expected 1 stale, got %d", report.Stale)
	}
	if report.Error != 1 {
		t.Errorf("expected 1 error, got %d", report.Error)
	}
}

func TestCheckHealth_TotalSyncsCount(t *testing.T) {
	entries := []Entry{
		makeHealthEntry("secret/app", "sync", 1),
		makeHealthEntry("secret/app", "sync", 2),
		makeHealthEntry("secret/app", "sync", 3),
	}
	report := CheckHealth(entries, 7)
	if len(report.Statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(report.Statuses))
	}
	if report.Statuses[0].TotalSyncs != 3 {
		t.Errorf("expected TotalSyncs=3, got %d", report.Statuses[0].TotalSyncs)
	}
}
