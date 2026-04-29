package audit_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourusername/vaultpull/internal/audit"
)

func makeQuotaEntry(action, path string, keys []string, when time.Time) audit.Entry {
	return audit.Entry{Action: action, Path: path, Keys: keys, Timestamp: when}
}

func TestSaveQuotaPolicy_NegativeValues(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveQuotaPolicy(dir, audit.QuotaPolicy{MaxSyncsPerHour: -1})
	if err == nil {
		t.Fatal("expected error for negative quota value")
	}
}

func TestSaveQuotaPolicy_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := audit.QuotaPolicy{MaxSyncsPerHour: 10, MaxKeysPerSync: 50, MaxPathsPerDay: 5}
	if err := audit.SaveQuotaPolicy(dir, p); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := audit.LoadQuotaPolicy(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got != p {
		t.Errorf("got %+v, want %+v", got, p)
	}
}

func TestLoadQuotaPolicy_NotFound(t *testing.T) {
	dir := t.TempDir()
	p, err := audit.LoadQuotaPolicy(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != (audit.QuotaPolicy{}) {
		t.Errorf("expected zero policy, got %+v", p)
	}
}

func TestSaveQuotaPolicy_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	if err := audit.SaveQuotaPolicy(dir, audit.QuotaPolicy{MaxSyncsPerHour: 5}); err != nil {
		t.Fatalf("save: %v", err)
	}
	info, err := os.Stat(dir + "/.vaultpull_quota.json")
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

func TestCheckQuota_NoViolations(t *testing.T) {
	now := time.Now().UTC()
	entries := []audit.Entry{
		makeQuotaEntry("sync", "secret/app", []string{"A", "B"}, now.Add(-10*time.Minute)),
	}
	p := audit.QuotaPolicy{MaxSyncsPerHour: 5, MaxKeysPerSync: 10, MaxPathsPerDay: 3}
	result := audit.CheckQuota(entries, p)
	if !result.OK() {
		t.Errorf("expected no violations, got %v", result.Violations)
	}
}

func TestCheckQuota_ExceedsSyncsPerHour(t *testing.T) {
	now := time.Now().UTC()
	var entries []audit.Entry
	for i := 0; i < 6; i++ {
		entries = append(entries, makeQuotaEntry("sync", "secret/app", []string{"K"}, now.Add(-time.Duration(i)*time.Minute)))
	}
	p := audit.QuotaPolicy{MaxSyncsPerHour: 5}
	result := audit.CheckQuota(entries, p)
	if result.OK() {
		t.Fatal("expected violation for syncs per hour")
	}
	if result.Violations[0].Rule != "max_syncs_per_hour" {
		t.Errorf("unexpected rule: %s", result.Violations[0].Rule)
	}
}

func TestCheckQuota_ExceedsKeysPerSync(t *testing.T) {
	now := time.Now().UTC()
	keys := make([]string, 20)
	for i := range keys {
		keys[i] = "KEY"
	}
	entries := []audit.Entry{makeQuotaEntry("sync", "secret/app", keys, now.Add(-5*time.Minute))}
	p := audit.QuotaPolicy{MaxKeysPerSync: 10}
	result := audit.CheckQuota(entries, p)
	if result.OK() {
		t.Fatal("expected violation for keys per sync")
	}
}

func TestCheckQuota_ExceedsPathsPerDay(t *testing.T) {
	now := time.Now().UTC()
	var entries []audit.Entry
	for _, path := range []string{"a", "b", "c", "d"} {
		entries = append(entries, makeQuotaEntry("sync", path, []string{"K"}, now.Add(-1*time.Hour)))
	}
	p := audit.QuotaPolicy{MaxPathsPerDay: 3}
	result := audit.CheckQuota(entries, p)
	if result.OK() {
		t.Fatal("expected violation for paths per day")
	}
}
