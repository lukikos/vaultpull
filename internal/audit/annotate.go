package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Annotation holds a user-defined note attached to a named tag or snapshot.
type Annotation struct {
	Name      string    `json:"name"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

func annotationFilePath(dir string) string {
	return filepath.Join(dir, ".vaultpull_annotations.json")
}

// SaveAnnotation writes a note for the given name into the annotations file.
// If an annotation with the same name already exists it is overwritten.
func SaveAnnotation(dir, name, note string) error {
	if name == "" {
		return fmt.Errorf("annotation name must not be empty")
	}

	annotations, _ := LoadAnnotations(dir)

	updated := false
	for i, a := range annotations {
		if a.Name == name {
			annotations[i].Note = note
			annotations[i].CreatedAt = time.Now().UTC()
			updated = true
			break
		}
	}
	if !updated {
		annotations = append(annotations, Annotation{
			Name:      name,
			Note:      note,
			CreatedAt: time.Now().UTC(),
		})
	}

	data, err := json.MarshalIndent(annotations, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal annotations: %w", err)
	}

	path := annotationFilePath(dir)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write annotations file: %w", err)
	}
	return nil
}

// LoadAnnotations reads all annotations from disk.
// Returns an empty slice when the file does not exist.
func LoadAnnotations(dir string) ([]Annotation, error) {
	path := annotationFilePath(dir)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Annotation{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read annotations file: %w", err)
	}

	var annotations []Annotation
	if err := json.Unmarshal(data, &annotations); err != nil {
		return nil, fmt.Errorf("parse annotations file: %w", err)
	}
	return annotations, nil
}
