package audit

import (
	"testing"
)

func makeRedactEntry(key, value string) Entry {
	return Entry{
		Key:   key,
		Value: value,
		Path:  "secret/app",
		Action: "write",
	}
}

func TestRedact_DefaultPatterns(t *testing.T) {
	entries := []Entry{
		makeRedactEntry("DB_PASSWORD", "s3cr3t"),
		makeRedactEntry("API_TOKEN", "tok-abc"),
		makeRedactEntry("APP_NAME", "myapp"),
	}

	result := Redact(entries, nil)

	if result[0].Value != "***" {
		t.Errorf("expected DB_PASSWORD to be redacted, got %q", result[0].Value)
	}
	if result[1].Value != "***" {
		t.Errorf("expected API_TOKEN to be redacted, got %q", result[1].Value)
	}
	if result[2].Value != "myapp" {
		t.Errorf("expected APP_NAME to be unchanged, got %q", result[2].Value)
	}
}

func TestRedact_CustomPatterns(t *testing.T) {
	entries := []Entry{
		makeRedactEntry("STRIPE_KEY", "sk_live_abc"),
		makeRedactEntry("REGION", "us-east-1"),
	}

	opts := &RedactOptions{Patterns: []string{"stripe"}, Mask: "[hidden]"}
	result := Redact(entries, opts)

	if result[0].Value != "[hidden]" {
		t.Errorf("expected STRIPE_KEY redacted with custom mask, got %q", result[0].Value)
	}
	if result[1].Value != "us-east-1" {
		t.Errorf("expected REGION unchanged, got %q", result[1].Value)
	}
}

func TestRedact_EmptyEntries(t *testing.T) {
	result := Redact([]Entry{}, nil)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d entries", len(result))
	}
}

func TestRedact_DoesNotMutateOriginal(t *testing.T) {
	original := []Entry{makeRedactEntry("DB_SECRET", "original")}
	_ = Redact(original, nil)
	if original[0].Value != "original" {
		t.Error("Redact must not mutate the original entries slice")
	}
}

func TestRedact_CaseInsensitiveKey(t *testing.T) {
	entries := []Entry{
		makeRedactEntry("db_password", "lower"),
		makeRedactEntry("DB_PASSWORD", "upper"),
		makeRedactEntry("Db_Password", "mixed"),
	}
	result := Redact(entries, nil)
	for _, e := range result {
		if e.Value != "***" {
			t.Errorf("expected %q to be redacted regardless of case", e.Key)
		}
	}
}
