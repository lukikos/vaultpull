package dotenv

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Writer writes secrets to a .env file.
type Writer struct {
	path   string
	backup bool
}

// NewWriter creates a new Writer for the given file path.
// If backup is true, an existing file is backed up before overwriting.
func NewWriter(path string, backup bool) *Writer {
	return &Writer{path: path, backup: backup}
}

// Write serialises the provided secrets map into KEY=VALUE lines and
// writes them to the configured file path, creating it if necessary.
func (w *Writer) Write(secrets map[string]string) error {
	if w.backup {
		if err := w.backupExisting(); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	var sb strings.Builder
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(&sb, "%s=%s\n", k, quoteValue(secrets[k]))
	}

	return os.WriteFile(w.path, []byte(sb.String()), 0600)
}

// backupExisting copies the existing file to <path>.bak if it exists.
func (w *Writer) backupExisting() error {
	data, err := os.ReadFile(w.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return os.WriteFile(w.path+".bak", data, 0600)
}

// quoteValue wraps the value in double quotes if it contains spaces or
// special characters that could break shell parsing.
func quoteValue(v string) string {
	if strings.ContainsAny(v, " \t\n#\"\'\\$") {
		v = strings.ReplaceAll(v, `"`, `\"`)
		return `"` + v + `"`
	}
	return v
}
