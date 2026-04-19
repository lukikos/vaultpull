package audit

import (
	"fmt"
	"os"
	"time"
)

// Rotate archives the current audit log by renaming it with a timestamp suffix,
// then starts fresh on the next write.
func Rotate(logPath string) (string, error) {
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return "", nil
	}

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	archivePath := fmt.Sprintf("%s.%s", logPath, timestamp)

	if err := os.Rename(logPath, archivePath); err != nil {
		return "", fmt.Errorf("rotate: rename failed: %w", err)
	}

	return archivePath, nil
}
