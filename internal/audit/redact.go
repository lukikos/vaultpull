package audit

import (
	"regexp"
	"strings"
)

// RedactOptions controls which keys are masked in output.
type RedactOptions struct {
	// Patterns is a list of glob-style substrings; any key containing one is redacted.
	Patterns []string
	// Mask is the replacement string. Defaults to "***".
	Mask string
}

var defaultSensitivePatterns = []string{
	"password", "secret", "token", "key", "apikey", "api_key",
	"passwd", "credential", "private", "auth",
}

// Redact returns a copy of entries with sensitive values replaced by a mask.
// If opts is nil, default sensitive patterns and mask "***" are used.
func Redact(entries []Entry, opts *RedactOptions) []Entry {
	patterns := defaultSensitivePatterns
	mask := "***"
	if opts != nil {
		if len(opts.Patterns) > 0 {
			patterns = opts.Patterns
		}
		if opts.Mask != "" {
			mask = opts.Mask
		}
	}

	out := make([]Entry, len(entries))
	for i, e := range entries {
		if isSensitive(e.Key, patterns) {
			e.Value = mask
		}
		out[i] = e
	}
	return out
}

// isSensitive returns true if key matches any of the given patterns (case-insensitive substring match).
func isSensitive(key string, patterns []string) bool {
	lower := strings.ToLower(key)
	for _, p := range patterns {
		matched, err := regexp.MatchString(strings.ToLower(p), lower)
		if err == nil && matched {
			return true
		}
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}
