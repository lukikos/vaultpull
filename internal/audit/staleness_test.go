package audit_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func makeStalenessEntry(path, action string, daysAgo int) audit.Entry {
	return audit.Entry{
		Timestamp: time.Now().UTC().Add(-time.Duration(daysAgo) * 24 * time.Hour),
		Path:      path,
		Action:    action,
	}
}

func TestCheckStaleness_Empty(t *testing.T) {
	reports := audit.CheckStaleness(nil, audit.StalenessOptions{ThresholdDays: 30})
	if len(reports) != 0 {
		t.Fatalf("expected 0 reports, got %d", len(reports))
	}
}

func TestCheckStaleness_FreshPath(t *testing.T) {
	entries := []audit.Entry{
		makeStalenessEntry("secret/app", "sync", 5),
	}
	reports := audit.CheckStaleness(entries, audit.StalenessOptions{ThresholdDays: 30})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].IsStale {
		t.Error("expected path to be fresh, got stale")
	}
	if reports[0].AgeDays < 4 || reports[0].AgeDays > 6 {
		t.Errorf("unexpected AgeDays: %d", reports[0].AgeDays)
	}
}

func TestCheckStaleness_StalePath(t *testing.T) {
	entries := []audit.Entry{
		makeStalenessEntry("secret/old", "sync", 45),
	}
	reports := audit.CheckStaleness(entries, audit.StalenessOptions{ThresholdDays: 30})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if !reports[0].IsStale {
		t.Error("expected path to be stale")
	}
}

func TestCheckStaleness_UsesLatestEntry(t *testing.T) {
	entries := []audit.Entry{
		makeStalenessEntry("secret/app", "sync", 60),
		makeStalenessEntry("secret/app", "sync", 3),
	}
	reports := audit.CheckStaleness(entries, audit.StalenessOptions{ThresholdDays: 30})
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].IsStale {
		t.Error("expected fresh path based on most recent sync")
	}
}

func TestCheckStaleness_DefaultThreshold(t *testing.T) {
	entries := []audit.Entry{
		makeStalenessEntry("secret/app", "sync", 35),
	}
	// ThresholdDays=0 should default to 30
	reports := audit.CheckStaleness(entries, audit.StalenessOptions{})
	if !reports[0].IsStale {
		t.Error("expected stale with default 30-day threshold")
	}
}

func TestCheckStaleness_IgnoresNonSyncActions(t *testing.T) {
	entries := []audit.Entry{
		makeStalenessEntry("secret/app", "read", 2),
	}
	reports := audit.CheckStaleness(entries, audit.StalenessOptions{ThresholdDays: 30})
	if len(reports) != 0 {
		t.Fatalf("expected 0 reports for non-sync actions, got %d", len(reports))
	}
}

func TestCheckStaleness_SortedByPath(t *testing.T) {
	entries := []audit.Entry{
		makeStalenessEntry("secret/z", "sync", 1),
		makeStalenessEntry("secret/a", "sync", 1),
		makeStalenessEntry("secret/m", "sync", 1),
	}
	reports := audit.CheckStaleness(entries, audit.StalenessOptions{ThresholdDays: 30})
	if reports[0].Path != "secret/a" || reports[1].Path != "secret/m" || reports[2].Path != "secret/z" {
		t.Errorf("reports not sorted by path: %v", reports)
	}
}
