package audit

import (
	"fmt"
	"testing"
	"time"
)

func makeExpiryEntry(path, key, value string, ts time.Time) Entry {
	return Entry{
		Timestamp: ts,
		Path:      path,
		Key:       key,
		Value:     value,
		Action:    "sync",
	}
}

func TestCheckExpiry_Empty(t *testing.T) {
	results := CheckExpiry(nil, DefaultExpiryConfig())
	if len(results) != 0 {
		t.Fatalf("expected no results, got %d", len(results))
	}
}

func TestCheckExpiry_NoExpiryValues(t *testing.T) {
	entries := []Entry{
		makeExpiryEntry("secret/app", "DB_PASS", "hunter2", time.Now()),
	}
	results := CheckExpiry(entries, DefaultExpiryConfig())
	if len(results) != 0 {
		t.Fatalf("expected 0 results for non-expiry values, got %d", len(results))
	}
}

func TestCheckExpiry_ExpiredKey(t *testing.T) {
	past := time.Now().UTC().Add(-48 * time.Hour)
	entries := []Entry{
		makeExpiryEntry("secret/app", "TOKEN", fmt.Sprintf("expires=%s", past.Format(time.RFC3339)), time.Now()),
	}
	results := CheckExpiry(entries, DefaultExpiryConfig())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Expired {
		t.Error("expected key to be marked expired")
	}
	if results[0].Key != "TOKEN" {
		t.Errorf("unexpected key: %s", results[0].Key)
	}
}

func TestCheckExpiry_WarningSoonKey(t *testing.T) {
	soon := time.Now().UTC().Add(10 * 24 * time.Hour)
	entries := []Entry{
		makeExpiryEntry("secret/app", "API_KEY", fmt.Sprintf("expires=%s", soon.Format(time.RFC3339)), time.Now()),
	}
	cfg := DefaultExpiryConfig() // WarnWithinDays = 30
	results := CheckExpiry(entries, cfg)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Expired {
		t.Error("key should not be expired, only expiring soon")
	}
	if results[0].DaysLeft < 9 || results[0].DaysLeft > 11 {
		t.Errorf("unexpected DaysLeft: %d", results[0].DaysLeft)
	}
}

func TestCheckExpiry_FreshKeyIgnored(t *testing.T) {
	far := time.Now().UTC().Add(90 * 24 * time.Hour)
	entries := []Entry{
		makeExpiryEntry("secret/app", "CERT", fmt.Sprintf("expires=%s", far.Format(time.RFC3339)), time.Now()),
	}
	results := CheckExpiry(entries, DefaultExpiryConfig())
	if len(results) != 0 {
		t.Fatalf("expected 0 results for fresh key, got %d", len(results))
	}
}

func TestCheckExpiry_UsesLatestEntry(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-48 * time.Hour)
	far := now.Add(90 * 24 * time.Hour)
	entries := []Entry{
		// older entry has expired value
		makeExpiryEntry("secret/app", "KEY", fmt.Sprintf("expires=%s", past.Format(time.RFC3339)), now.Add(-1*time.Hour)),
		// newer entry has fresh value
		makeExpiryEntry("secret/app", "KEY", fmt.Sprintf("expires=%s", far.Format(time.RFC3339)), now),
	}
	results := CheckExpiry(entries, DefaultExpiryConfig())
	if len(results) != 0 {
		t.Fatalf("expected 0 results (latest entry is fresh), got %d", len(results))
	}
}
