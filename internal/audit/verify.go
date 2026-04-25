package audit

import (
	"fmt"
	"os"
	"strings"
)

// VerifyResult holds the outcome of verifying a .env file against audit log entries.
type VerifyResult struct {
	Path        string
	MissingKeys []string
	ExtraKeys   []string
	MatchedKeys []string
	OK          bool
}

// Verify checks whether the current .env file at envPath matches the most
// recently synced keys recorded in the audit log at logPath for that secret path.
func Verify(logPath, secretPath, envPath string) (VerifyResult, error) {
	result := VerifyResult{Path: secretPath}

	entries, err := ReadAll(logPath)
	if err != nil {
		return result, fmt.Errorf("reading audit log: %w", err)
	}

	// Replay the latest state for secretPath from the audit log.
	expected, err := Replay(logPath, secretPath, entries[len(entries)-1].Timestamp.Add(1))
	if err != nil {
		return result, fmt.Errorf("replaying audit log: %w", err)
	}

	// Parse the actual .env file.
	data, err := os.ReadFile(envPath)
	if os.IsNotExist(err) {
		for k := range expected {
			result.MissingKeys = append(result.MissingKeys, k)
		}
		return result, nil
	}
	if err != nil {
		return result, fmt.Errorf("reading env file: %w", err)
	}

	actual := parseEnvLines(string(data))

	for k := range expected {
		if _, ok := actual[k]; ok {
			result.MatchedKeys = append(result.MatchedKeys, k)
		} else {
			result.MissingKeys = append(result.MissingKeys, k)
		}
	}
	for k := range actual {
		if _, ok := expected[k]; !ok {
			result.ExtraKeys = append(result.ExtraKeys, k)
		}
	}

	result.OK = len(result.MissingKeys) == 0
	return result, nil
}

// parseEnvLines parses KEY=VALUE lines from a .env file string.
func parseEnvLines(content string) map[string]string {
	m := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return m
}
