package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SchemaVersion represents a versioned schema definition for audit log entries.
type SchemaVersion struct {
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Fields    []string  `json:"fields"`
}

// SchemaValidationError holds a list of field-level issues found during validation.
type SchemaValidationError struct {
	Issues []string
}

func (e *SchemaValidationError) Error() string {
	return fmt.Sprintf("schema validation failed: %d issue(s)", len(e.Issues))
}

func schemaFilePath(dir string) string {
	return filepath.Join(dir, ".vaultpull_schema.json")
}

// SaveSchema writes a SchemaVersion record to the audit directory.
func SaveSchema(dir string, sv SchemaVersion) error {
	if sv.Version <= 0 {
		return fmt.Errorf("schema version must be a positive integer")
	}
	sv.CreatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(sv, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}
	path := schemaFilePath(dir)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write schema file: %w", err)
	}
	return nil
}

// LoadSchema reads the SchemaVersion from the audit directory.
func LoadSchema(dir string) (SchemaVersion, error) {
	path := schemaFilePath(dir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return SchemaVersion{}, fmt.Errorf("schema file not found")
		}
		return SchemaVersion{}, fmt.Errorf("read schema file: %w", err)
	}
	var sv SchemaVersion
	if err := json.Unmarshal(data, &sv); err != nil {
		return SchemaVersion{}, fmt.Errorf("unmarshal schema: %w", err)
	}
	return sv, nil
}

// ValidateEntries checks that each entry contains all required fields from the schema.
func ValidateEntries(entries []LogEntry, sv SchemaVersion) *SchemaValidationError {
	required := map[string]bool{}
	for _, f := range sv.Fields {
		required[f] = true
	}
	var issues []string
	for i, e := range entries {
		if required["key"] && e.Key == "" {
			issues = append(issues, fmt.Sprintf("entry %d: missing field 'key'", i))
		}
		if required["path"] && e.Path == "" {
			issues = append(issues, fmt.Sprintf("entry %d: missing field 'path'", i))
		}
		if required["action"] && e.Action == "" {
			issues = append(issues, fmt.Sprintf("entry %d: missing field 'action'", i))
		}
	}
	if len(issues) == 0 {
		return nil
	}
	return &SchemaValidationError{Issues: issues}
}
