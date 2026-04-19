package audit

import (
	"bufio"
	"encoding/json"
	"os"
)

// ReadAll reads all audit log entries from the given file path.
func ReadAll(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, sc.Err()
}
