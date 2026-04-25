package audit

import (
	"fmt"
	"strings"
)

// LintResult holds a single lint warning for an audit log entry.
type LintResult struct {
	Index   int
	Path    string
	Key     string
	Warning string
}

// LintReport holds all warnings produced by a Lint run.
type LintReport struct {
	Warnings []LintResult
	Total    int
	Clean    bool
}

// Lint inspects audit log entries for common issues such as empty keys,
// suspicious key names, or unknown actions.
func Lint(logPath string) (LintReport, error) {
	entries, err := ReadAll(logPath)
	if err != nil {
		return LintReport{}, fmt.Errorf("lint: read log: %w", err)
	}

	validActions := map[string]bool{
		"added":   true,
		"updated": true,
		"removed": true,
		"synced":  true,
	}

	var warnings []LintResult

	for i, e := range entries {
		if strings.TrimSpace(e.Key) == "" {
			warnings = append(warnings, LintResult{
				Index:   i,
				Path:    e.Path,
				Key:     e.Key,
				Warning: "empty key",
			})
		}

		if strings.TrimSpace(e.Path) == "" {
			warnings = append(warnings, LintResult{
				Index:   i,
				Path:    e.Path,
				Key:     e.Key,
				Warning: "empty path",
			})
		}

		if !validActions[strings.ToLower(e.Action)] {
			warnings = append(warnings, LintResult{
				Index:   i,
				Path:    e.Path,
				Key:     e.Key,
				Warning: fmt.Sprintf("unknown action %q", e.Action),
			})
		}

		if strings.Contains(e.Key, " ") {
			warnings = append(warnings, LintResult{
				Index:   i,
				Path:    e.Path,
				Key:     e.Key,
				Warning: "key contains whitespace",
			})
		}
	}

	return LintReport{
		Warnings: warnings,
		Total:    len(warnings),
		Clean:    len(warnings) == 0,
	}, nil
}
