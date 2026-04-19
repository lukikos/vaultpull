package dotenv

import (
	"bufio"
	"os"
	"strings"
)

// Merge reads an existing .env file and merges incoming secrets into it.
// Existing keys are overwritten; keys not present in secrets are preserved.
func Merge(path string, secrets map[string]string) (map[string]string, error) {
	existing, err := parse(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if existing == nil {
		existing = make(map[string]string)
	}
	for k, v := range secrets {
		existing[k] = v
	}
	return existing, nil
}

// parse reads a simple KEY=VALUE .env file into a map.
// Lines beginning with # and blank lines are ignored.
func parse(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		result[key] = val
	}
	return result, scanner.Err()
}
