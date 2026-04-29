package audit

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

// Pin represents a named, immutable marker attached to a specific audit entry.
type Pin struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	Note      string    `json:"note,omitempty"`
}

func pinFilePath(dir string) string {
	return filepath.Join(dir, "pins.json")
}

// SavePin persists a named pin to the pin store in dir.
func SavePin(dir, name, path, key, value, note string) error {
	if name == "" {
		return errors.New("pin name must not be empty")
	}
	if path == "" {
		return errors.New("pin path must not be empty")
	}
	if key == "" {
		return errors.New("pin key must not be empty")
	}

	pins, _ := LoadPins(dir)

	pin := Pin{
		Name:      name,
		Path:      path,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now().UTC(),
		Note:      note,
	}

	// Replace existing pin with same name if present.
	updated := false
	for i, p := range pins {
		if p.Name == name {
			pins[i] = pin
			updated = true
			break
		}
	}
	if !updated {
		pins = append(pins, pin)
	}

	data, err := json.MarshalIndent(pins, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pinFilePath(dir), data, 0o600)
}

// LoadPins reads all saved pins from dir. Returns an empty slice if none exist.
func LoadPins(dir string) ([]Pin, error) {
	data, err := os.ReadFile(pinFilePath(dir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Pin{}, nil
		}
		return nil, err
	}
	var pins []Pin
	if err := json.Unmarshal(data, &pins); err != nil {
		return nil, err
	}
	return pins, nil
}
