package audit

import (
	"encoding/json"
	"os"
	"time"
)

// Entry represents a single audit log entry.
type Entry struct {
	Timestamp  time.Time `json:"timestamp"`
	SecretPath string    `json:"secret_path"`
	Keys       []string  `json:"keys"`
	OutputFile string    `json:"output_file"`
	Status     string    `json:"status"`
	Error      string    `json:"error,omitempty"`
}

// Logger writes audit entries to a file.
type Logger struct {
	path string
}

// NewLogger creates a Logger that appends to the given file path.
func NewLogger(path string) *Logger {
	return &Logger{path: path}
}

// Record appends an Entry to the audit log file as a JSON line.
func (l *Logger) Record(e Entry) error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	e.Timestamp = time.Now().UTC()
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return err
}
