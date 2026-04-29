package audit

import (
	"testing"
	"time"
)

func makeBadgeEntry(path, action string, ts time.Time) Entry {
	return Entry{Path: path, Action: action, Timestamp: ts}
}

func TestGenerateBadges_Empty(t *testing.T) {
	badges := GenerateBadges(nil, DefaultBadgeConfig())
	if len(badges) != 0 {
		t.Fatalf("expected 0 badges, got %d", len(badges))
	}
}

func TestGenerateBadges_OKStatus(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeBadgeEntry("secret/app", "sync", now.Add(-1*time.Hour)),
		makeBadgeEntry("secret/app", "sync", now.Add(-30*time.Minute)),
	}
	badges := GenerateBadges(entries, DefaultBadgeConfig())
	if len(badges) != 1 {
		t.Fatalf("expected 1 badge, got %d", len(badges))
	}
	if badges[0].Status != BadgeStatusOK {
		t.Errorf("expected ok, got %s", badges[0].Status)
	}
	if badges[0].LastSync == nil {
		t.Error("expected LastSync to be set")
	}
}

func TestGenerateBadges_StaleStatus(t *testing.T) {
	old := time.Now().Add(-48 * time.Hour)
	entries := []Entry{
		makeBadgeEntry("secret/app", "sync", old),
	}
	badges := GenerateBadges(entries, DefaultBadgeConfig())
	if len(badges) != 1 {
		t.Fatalf("expected 1 badge, got %d", len(badges))
	}
	if badges[0].Status != BadgeStatusWarning {
		t.Errorf("expected warning, got %s", badges[0].Status)
	}
	if badges[0].Message != "stale" {
		t.Errorf("expected message 'stale', got %s", badges[0].Message)
	}
}

func TestGenerateBadges_ErrorStatus(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeBadgeEntry("secret/app", "error", now.Add(-5*time.Minute)),
		makeBadgeEntry("secret/app", "error", now.Add(-4*time.Minute)),
		makeBadgeEntry("secret/app", "sync", now.Add(-3*time.Minute)),
	}
	cfg := DefaultBadgeConfig()
	cfg.ErrorRateThreshold = 0.5
	badges := GenerateBadges(entries, cfg)
	if len(badges) != 1 {
		t.Fatalf("expected 1 badge, got %d", len(badges))
	}
	if badges[0].Status != BadgeStatusError {
		t.Errorf("expected error, got %s", badges[0].Status)
	}
}

func TestGenerateBadges_MultiplePaths(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeBadgeEntry("secret/a", "sync", now.Add(-1*time.Hour)),
		makeBadgeEntry("secret/b", "sync", now.Add(-2*time.Hour)),
	}
	badges := GenerateBadges(entries, DefaultBadgeConfig())
	if len(badges) != 2 {
		t.Fatalf("expected 2 badges, got %d", len(badges))
	}
}
