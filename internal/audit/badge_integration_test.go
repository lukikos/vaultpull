package audit

import (
	"testing"
	"time"
)

func TestBadge_AfterWriteIsOK(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	if err := logger.Record("secret/integration", "sync", []string{"TOKEN", "DB_PASS"}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	entries, err := ReadAll(dir)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}

	badges := GenerateBadges(entries, DefaultBadgeConfig())
	if len(badges) != 1 {
		t.Fatalf("expected 1 badge, got %d", len(badges))
	}
	if badges[0].Status != BadgeStatusOK {
		t.Errorf("expected ok status after fresh sync, got %s", badges[0].Status)
	}
	if badges[0].Path != "secret/integration" {
		t.Errorf("unexpected path: %s", badges[0].Path)
	}
	if badges[0].LastSync == nil {
		t.Error("expected LastSync to be populated")
	}
}

func TestBadge_ErrorEntriesTriggerErrorStatus(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewLogger(dir)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	for i := 0; i < 4; i++ {
		if err := logger.Record("secret/flaky", "error", []string{"KEY"}); err != nil {
			t.Fatalf("Record error: %v", err)
		}
	}
	if err := logger.Record("secret/flaky", "sync", []string{"KEY"}); err != nil {
		t.Fatalf("Record sync: %v", err)
	}

	entries, _ := ReadAll(dir)
	cfg := DefaultBadgeConfig()
	cfg.ErrorRateThreshold = 0.5
	badges := GenerateBadges(entries, cfg)

	if len(badges) != 1 {
		t.Fatalf("expected 1 badge, got %d", len(badges))
	}
	if badges[0].Status != BadgeStatusError {
		t.Errorf("expected error status, got %s", badges[0].Status)
	}
}

func TestBadge_GeneratedAtIsRecent(t *testing.T) {
	dir := t.TempDir()
	logger, _ := NewLogger(dir)
	_ = logger.Record("secret/ts", "sync", []string{"K"})

	entries, _ := ReadAll(dir)
	before := time.Now()
	badges := GenerateBadges(entries, DefaultBadgeConfig())
	after := time.Now()

	for _, b := range badges {
		if b.GeneratedAt.Before(before) || b.GeneratedAt.After(after) {
			t.Errorf("GeneratedAt %v out of range [%v, %v]", b.GeneratedAt, before, after)
		}
	}
}
