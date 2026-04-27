package audit

import (
	"testing"
	"time"
)

func makeAlertEntry(path, action string, ts time.Time) Entry {
	return Entry{Path: path, Action: action, Timestamp: ts}
}

func TestCheckAlerts_Empty(t *testing.T) {
	alerts := CheckAlerts(nil, DefaultAlertConfig())
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestCheckAlerts_NoIssues(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeAlertEntry("secret/app", "sync", now.Add(-1*time.Hour)),
		makeAlertEntry("secret/app", "sync", now.Add(-30*time.Minute)),
	}
	alerts := CheckAlerts(entries, DefaultAlertConfig())
	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d: %+v", len(alerts), alerts)
	}
}

func TestCheckAlerts_HighErrorRate(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeAlertEntry("secret/app", "sync", now.Add(-1*time.Hour)),
		makeAlertEntry("secret/app", "error", now.Add(-50*time.Minute)),
		makeAlertEntry("secret/app", "error", now.Add(-40*time.Minute)),
		makeAlertEntry("secret/app", "error", now.Add(-30*time.Minute)),
	}
	cfg := DefaultAlertConfig()
	cfg.MaxErrorRate = 0.1
	alerts := CheckAlerts(entries, cfg)

	var found bool
	for _, a := range alerts {
		if a.Path == "secret/app" && a.Level == AlertCritical {
			found = true
		}
	}
	if !found {
		t.Error("expected critical alert for high error rate")
	}
}

func TestCheckAlerts_StalePath(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeAlertEntry("secret/db", "sync", now.Add(-72*time.Hour)),
	}
	cfg := DefaultAlertConfig()
	alerts := CheckAlerts(entries, cfg)

	var found bool
	for _, a := range alerts {
		if a.Path == "secret/db" && a.Level == AlertCritical {
			found = true
		}
	}
	if !found {
		t.Error("expected critical stale alert for secret/db")
	}
}

func TestCheckAlerts_WarningSyncInterval(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeAlertEntry("secret/svc", "sync", now.Add(-30*time.Hour)),
	}
	cfg := DefaultAlertConfig()
	// 30h > MaxSyncInterval(24h) but < StaleAfter(48h)
	alerts := CheckAlerts(entries, cfg)

	var found bool
	for _, a := range alerts {
		if a.Path == "secret/svc" && a.Level == AlertWarning {
			found = true
		}
	}
	if !found {
		t.Error("expected warning alert for overdue sync on secret/svc")
	}
}

func TestCheckAlerts_MultiplePathsIndependent(t *testing.T) {
	now := time.Now()
	entries := []Entry{
		makeAlertEntry("secret/a", "sync", now.Add(-1*time.Hour)),
		makeAlertEntry("secret/b", "sync", now.Add(-72*time.Hour)),
	}
	alerts := CheckAlerts(entries, DefaultAlertConfig())

	for _, a := range alerts {
		if a.Path == "secret/a" {
			t.Errorf("unexpected alert for secret/a: %+v", a)
		}
	}
	var foundB bool
	for _, a := range alerts {
		if a.Path == "secret/b" {
			foundB = true
		}
	}
	if !foundB {
		t.Error("expected alert for stale secret/b")
	}
}
