package audit_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultpull/internal/audit"
)

func TestSaveTTLPolicy_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveTTLPolicy(dir, "", time.Hour)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSaveTTLPolicy_ZeroTTL(t *testing.T) {
	dir := t.TempDir()
	err := audit.SaveTTLPolicy(dir, "secret/app", 0)
	if err == nil {
		t.Fatal("expected error for zero TTL")
	}
}

func TestSaveTTLPolicy_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := audit.SaveTTLPolicy(dir, "secret/app", 2*time.Hour); err != nil {
		t.Fatalf("SaveTTLPolicy: %v", err)
	}
	policies, err := audit.LoadTTLPolicies(dir)
	if err != nil {
		t.Fatalf("LoadTTLPolicies: %v", err)
	}
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(policies))
	}
	if policies[0].Path != "secret/app" {
		t.Errorf("unexpected path: %s", policies[0].Path)
	}
	if policies[0].TTL != 2*time.Hour {
		t.Errorf("unexpected TTL: %v", policies[0].TTL)
	}
}

func TestLoadTTLPolicies_NotFound(t *testing.T) {
	dir := t.TempDir()
	policies, err := audit.LoadTTLPolicies(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policies) != 0 {
		t.Fatalf("expected empty slice, got %d", len(policies))
	}
}

func TestCheckTTL_FreshPath(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SaveTTLPolicy(dir, "secret/app", 24*time.Hour)

	entries := []audit.Entry{
		{Path: "secret/app", Action: "sync", Timestamp: time.Now().UTC().Add(-1 * time.Hour)},
	}
	results, err := audit.CheckTTL(dir, entries)
	if err != nil {
		t.Fatalf("CheckTTL: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Expired {
		t.Error("expected path to be fresh, got expired")
	}
}

func TestCheckTTL_ExpiredPath(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SaveTTLPolicy(dir, "secret/app", time.Minute)

	entries := []audit.Entry{
		{Path: "secret/app", Action: "sync", Timestamp: time.Now().UTC().Add(-2 * time.Hour)},
	}
	results, err := audit.CheckTTL(dir, entries)
	if err != nil {
		t.Fatalf("CheckTTL: %v", err)
	}
	if !results[0].Expired {
		t.Error("expected path to be expired")
	}
}

func TestCheckTTL_NoSyncEntry(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SaveTTLPolicy(dir, "secret/app", time.Hour)

	results, err := audit.CheckTTL(dir, []audit.Entry{})
	if err != nil {
		t.Fatalf("CheckTTL: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Expired {
		t.Error("expected expired when no sync entry exists")
	}
}
