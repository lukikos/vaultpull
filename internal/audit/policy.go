package audit

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// PolicyRule defines a rule that restricts which secret paths are allowed.
type PolicyRule struct {
	Path    string `json:"path"`
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// Policy holds a named set of rules.
type Policy struct {
	Name      string       `json:"name"`
	CreatedAt time.Time    `json:"created_at"`
	Rules     []PolicyRule `json:"rules"`
}

// PolicyViolation describes a path that violated a policy.
type PolicyViolation struct {
	Path   string
	Reason string
}

func policyFilePath(dir, name string) string {
	return filepath.Join(dir, ".vaultpull_policy_"+name+".json")
}

// SavePolicy persists a named policy to the given directory.
func SavePolicy(dir, name string, rules []PolicyRule) error {
	if name == "" {
		return errors.New("policy name must not be empty")
	}
	p := Policy{
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Rules:     rules,
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(policyFilePath(dir, name), data, 0600)
}

// LoadPolicy loads a named policy from the given directory.
func LoadPolicy(dir, name string) (Policy, error) {
	if name == "" {
		return Policy{}, errors.New("policy name must not be empty")
	}
	data, err := os.ReadFile(policyFilePath(dir, name))
	if err != nil {
		return Policy{}, err
	}
	var p Policy
	if err := json.Unmarshal(data, &p); err != nil {
		return Policy{}, err
	}
	return p, nil
}

// EnforcePolicy checks a list of paths against a policy and returns violations.
func EnforcePolicy(p Policy, paths []string) []PolicyViolation {
	ruleMap := make(map[string]PolicyRule, len(p.Rules))
	for _, r := range p.Rules {
		ruleMap[r.Path] = r
	}
	var violations []PolicyViolation
	for _, path := range paths {
		if rule, ok := ruleMap[path]; ok && !rule.Allowed {
			reason := rule.Reason
			if reason == "" {
				reason = "denied by policy"
			}
			violations = append(violations, PolicyViolation{Path: path, Reason: reason})
		}
	}
	return violations
}
