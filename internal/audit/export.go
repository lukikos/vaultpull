package audit

import (
	"encoding/csv"
	"fmt"
	"io"
	"time"
)

// ExportCSV writes audit log entries to w in CSV format.
func ExportCSV(entries []Entry, w io.Writer) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{"timestamp", "path", "keys_written", "output_file", "status"}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("writing csv header: %w", err)
	}

	for _, e := range entries {
		row := []string{
			e.Timestamp.Format(time.RFC3339),
			e.Path,
			fmt.Sprintf("%d", e.KeysWritten),
			e.OutputFile,
			e.Status,
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("writing csv row: %w", err)
		}
	}

	return cw.Error()
}
