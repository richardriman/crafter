// Package claudesettings provides read/mutate/write primitives for Claude
// Code's settings.json. It is the Go counterpart of the tolerant `node -e`
// blocks in install.sh, which load a settings file (treating a missing or
// garbled file as an empty object), mutate a few well-known keys, and write
// back `JSON.stringify(settings, null, 2) + "\n"`.
//
// Settings are held as a map of raw JSON values so that arbitrary, unknown
// keys (statusLine, hooks, model, permissions, env, ...) round-trip intact:
// only the keys a caller explicitly touches are ever changed.
//
// Key-ordering note: Go's encoding/json marshals map keys in alphabetical
// order, whereas Node's JSON.stringify preserves insertion order. This writer
// therefore emits keys alphabetically, which can differ from the byte order a
// prior Node run produced. JSON key order is not semantically meaningful, so
// this is acceptable for fresh/small settings; downstream phases that compare
// against exact Node-era byte output should be aware of this difference.
package claudesettings

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Settings is an in-memory view of a settings.json object. Values are kept as
// raw JSON so that keys the caller does not touch are preserved verbatim on
// write.
type Settings struct {
	values map[string]json.RawMessage
}

// New returns an empty Settings object.
func New() *Settings {
	return &Settings{values: map[string]json.RawMessage{}}
}

// Load reads and parses the settings file at path.
//
// Mirroring the tolerant posture of the install.sh node blocks
// (`try { settings = JSON.parse(...) } catch (e) {}`), a missing file OR
// content that is not valid JSON degrades to empty Settings with no error.
// Only a genuine, non-"not exist" I/O error (e.g. a permission problem) is
// surfaced to the caller.
func Load(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return New(), nil
		}
		return nil, fmt.Errorf("reading settings: %w", err)
	}

	values := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &values); err != nil {
		// Invalid / garbled JSON is treated as empty settings, just like the
		// node blocks' catch-and-default-to-{} behaviour.
		return New(), nil
	}
	return &Settings{values: values}, nil
}

// Has reports whether key is present in the settings.
func (s *Settings) Has(key string) bool {
	_, ok := s.values[key]
	return ok
}

// Get returns the raw JSON value stored under key and whether it was present.
func (s *Settings) Get(key string) (json.RawMessage, bool) {
	v, ok := s.values[key]
	return v, ok
}

// Set marshals value and stores it under key, replacing any existing value.
func (s *Settings) Set(key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshalling settings value for %q: %w", key, err)
	}
	s.values[key] = raw
	return nil
}

// SetRaw stores an already-encoded JSON value under key.
func (s *Settings) SetRaw(key string, value json.RawMessage) {
	s.values[key] = value
}

// Delete removes key from the settings.
func (s *Settings) Delete(key string) {
	delete(s.values, key)
}

// Marshal renders the settings as 2-space-indented JSON with a trailing
// newline, matching the byte shape of `JSON.stringify(settings, null, 2) + "\n"`.
func (s *Settings) Marshal() ([]byte, error) {
	data, err := json.MarshalIndent(s.values, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshalling settings: %w", err)
	}
	return append(data, '\n'), nil
}

// Save atomically writes the settings to path as 2-space-indented JSON with a
// trailing newline, via a temp file + rename (mirroring the skillbook store).
func Save(path string, s *Settings) error {
	data, err := s.Marshal()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating settings directory: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing settings temp file: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("renaming settings temp file: %w", err)
	}
	return nil
}

// Backup copies the settings file at path to path+".bak" before a destructive
// overwrite.
//
// Hygiene scheme: a pre-existing .bak is never clobbered. The first backup
// captured for a settings file is the one preserved, on the assumption that it
// holds the user's original (foreign) value that must not be destroyed by a
// later overwrite. This makes Backup idempotent and safe to re-run: the second
// and subsequent calls are no-ops once a .bak exists.
//
// If the settings file itself does not yet exist there is nothing to back up,
// so Backup returns nil without creating a .bak.
func Backup(path string) error {
	bak := path + ".bak"

	if _, err := os.Stat(bak); err == nil {
		// A backup already exists; preserve it untouched.
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking settings backup: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Nothing to back up yet.
			return nil
		}
		return fmt.Errorf("reading settings for backup: %w", err)
	}

	if err := os.WriteFile(bak, data, 0o644); err != nil {
		return fmt.Errorf("writing settings backup: %w", err)
	}
	return nil
}

// BackupForOverwrite saves rawBytes (the raw on-disk content of the settings
// file, read BEFORE any Load/mutation) as a recovery artifact before a
// destructive foreign-overwrite.
//
// This function addresses two deferred Phase-1 review findings:
//
//   - Finding #2: the backup is taken from the raw on-disk bytes, not from a
//     re-serialized Settings object. This means a malformed-but-meaningful
//     original file remains recoverable even when Load treated it as empty.
//
//   - Finding #3: the backup is content-aware so that a stale, unrelated .bak
//     cannot shadow the foreign value being destroyed:
//   - No .bak exists: write rawBytes to path+".bak".
//   - .bak exists AND its content equals rawBytes: skip — the overwrite was
//     already backed up (idempotent re-run is safe).
//   - .bak exists AND its content DIFFERS from rawBytes: the existing .bak holds
//     a different backup; write rawBytes to path+".bak.1" (or ".bak.2", etc.,
//     choosing the lowest suffix whose target does not yet exist).
//
// The caller must pass the raw bytes read from path before calling Load, because
// Load is tolerant and silently treats garbled content as an empty object —
// re-serialising the loaded Settings would lose the original on-disk bytes.
func BackupForOverwrite(path string, rawBytes []byte) error {
	bak := path + ".bak"

	existing, err := os.ReadFile(bak)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking settings backup: %w", err)
	}

	if err == nil {
		// A .bak already exists.
		if bytes.Equal(existing, rawBytes) {
			// Content matches — this exact foreign value is already safely backed
			// up. Re-run is a no-op.
			return nil
		}
		// The existing .bak holds different content (a prior backup, an unrelated
		// value). Find the next available numbered sibling so the current foreign
		// value is still preserved.
		const maxBackups = 4096
		for n := 1; n <= maxBackups; n++ {
			numbered := fmt.Sprintf("%s.bak.%d", path, n)
			f, err := os.OpenFile(numbered, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err == nil {
				_, werr := f.Write(rawBytes)
				cerr := f.Close()
				if werr != nil {
					return fmt.Errorf("writing numbered settings backup %s: %w", numbered, werr)
				}
				if cerr != nil {
					return fmt.Errorf("closing numbered settings backup %s: %w", numbered, cerr)
				}
				return nil
			}
			if !errors.Is(err, os.ErrExist) {
				return fmt.Errorf("checking numbered backup %s: %w", numbered, err)
			}
			// That numbered slot is already taken; try the next.
		}
		return fmt.Errorf("too many backup files for %s: all .bak.1 through .bak.%d already exist", path, maxBackups)
	}

	// No .bak at all: create path+".bak".
	if err := os.WriteFile(bak, rawBytes, 0o644); err != nil {
		return fmt.Errorf("writing settings backup: %w", err)
	}
	return nil
}
