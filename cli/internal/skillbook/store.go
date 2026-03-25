package skillbook

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// NewSkillbook returns an empty skillbook with version 1.
func NewSkillbook() *Skillbook {
	return &Skillbook{
		Version: 1,
		Skills:  []Skill{},
	}
}

// Load reads and unmarshals a skillbook from path.
// If the file does not exist, it returns an empty skillbook.
func Load(path string) (*Skillbook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewSkillbook(), nil
		}
		return nil, fmt.Errorf("reading skillbook: %w", err)
	}

	var sb Skillbook
	if err := json.Unmarshal(data, &sb); err != nil {
		return nil, fmt.Errorf("parsing skillbook: %w", err)
	}
	return &sb, nil
}

// Save atomically writes sb to path as indented JSON.
// It updates UpdatedAt to the current UTC time before writing.
func Save(path string, sb *Skillbook) error {
	sb.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	data, err := json.MarshalIndent(sb, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling skillbook: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing skillbook temp file: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("renaming skillbook temp file: %w", err)
	}
	return nil
}

// NewID generates a random hex ID using crypto/rand.
func NewID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating id: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}
