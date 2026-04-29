package audit

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// Watermark records the high-water mark (latest sync timestamp) for each
// secret path so consumers can quickly determine how fresh their data is.
type Watermark struct {
	Path      string    `json:"path"`
	LatestAt  time.Time `json:"latest_at"`
	SyncCount int       `json:"sync_count"`
}

func watermarkFilePath(dir string) string {
	return filepath.Join(dir, ".vaultpull_watermarks.json")
}

// ComputeWatermarks scans entries and returns the latest sync timestamp per path.
func ComputeWatermarks(entries []Entry) []Watermark {
	type agg struct {
		latest time.Time
		count  int
	}
	m := make(map[string]*agg)
	for _, e := range entries {
		if e.Action != "sync" {
			continue
		}
		a, ok := m[e.Path]
		if !ok {
			a = &agg{}
			m[e.Path] = a
		}
		a.count++
		if e.Timestamp.After(a.latest) {
			a.latest = e.Timestamp
		}
	}
	out := make([]Watermark, 0, len(m))
	for path, a := range m {
		out = append(out, Watermark{Path: path, LatestAt: a.latest, SyncCount: a.count})
	}
	return out
}

// SaveWatermarks persists watermarks to dir.
func SaveWatermarks(dir string, wm []Watermark) error {
	data, err := json.MarshalIndent(wm, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(watermarkFilePath(dir), data, 0o600)
}

// LoadWatermarks reads previously saved watermarks from dir.
func LoadWatermarks(dir string) ([]Watermark, error) {
	data, err := os.ReadFile(watermarkFilePath(dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var wm []Watermark
	if err := json.Unmarshal(data, &wm); err != nil {
		return nil, err
	}
	return wm, nil
}
