package audit_test

import (
	"os"
	"testing"

	"github.com/your-org/vaultpull/internal/audit"
)

func TestSavePolicy_EmptyName(t *testing.T) {
	dir := t.TempDir()
	err := audit.SavePolicy(dir, "", nil)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSavePolicy_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	rules := []audit.PolicyRule{
		{Path: "secret/prod", Allowed: false, Reason: "prod restricted"},
	}
	if err := audit.SavePolicy(dir, "default", rules); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}
}

func TestSavePolicy_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	_ = audit.SavePolicy(dir, "perms", nil)
	entries, _ := os.ReadDir(dir)
	if len(entries) == 0 {
		t.Fatal("no file created")
	}
	info, _ := entries[0].Info()
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

func TestLoadPolicy_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.LoadPolicy(dir, "missing")
	if err == nil {
		t.Fatal("expected error for missing policy")
	}
}

func TestLoadPolicy_EmptyName(t *testing.T) {
	dir := t.TempDir()
	_, err := audit.LoadPolicy(dir, "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestLoadPolicy_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	rules := []audit.PolicyRule{
		{Path: "secret/dev", Allowed: true},
		{Path: "secret/prod", Allowed: false, Reason: "restricted"},
	}
	if err := audit.SavePolicy(dir, "myp", rules); err != nil {
		t.Fatalf("save: %v", err)
	}
	p, err := audit.LoadPolicy(dir, "myp")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p.Name != "myp" {
		t.Errorf("expected name myp, got %s", p.Name)
	}
	if len(p.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(p.Rules))
	}
}

func TestEnforcePolicy_NoViolations(t *testing.T) {
	p := audit.Policy{
		Rules: []audit.PolicyRule{
			{Path: "secret/prod", Allowed: false},
		},
	}
	v := audit.EnforcePolicy(p, []string{"secret/dev"})
	if len(v) != 0 {
		t.Errorf("expected no violations, got %d", len(v))
	}
}

func TestEnforcePolicy_DetectsViolation(t *testing.T) {
	p := audit.Policy{
		Rules: []audit.PolicyRule{
			{Path: "secret/prod", Allowed: false, Reason: "prod restricted"},
		},
	}
	v := audit.EnforcePolicy(p, []string{"secret/prod", "secret/dev"})
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Path != "secret/prod" {
		t.Errorf("unexpected path %s", v[0].Path)
	}
	if v[0].Reason != "prod restricted" {
		t.Errorf("unexpected reason %s", v[0].Reason)
	}
}

func TestEnforcePolicy_DefaultReason(t *testing.T) {
	p := audit.Policy{
		Rules: []audit.PolicyRule{
			{Path: "secret/x", Allowed: false},
		},
	}
	v := audit.EnforcePolicy(p, []string{"secret/x"})
	if len(v) == 0 {
		t.Fatal("expected violation")
	}
	if v[0].Reason == "" {
		t.Error("expected non-empty default reason")
	}
}
