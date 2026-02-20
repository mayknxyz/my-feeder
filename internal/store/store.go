// Package store handles reading and writing flat-file storage: the article
// cache (ephemeral JSON), read state (synced JSON), and bookmarks (synced
// markdown). No database — just files and encoding/json.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// State tracks which articles have been read. This file is designed to be
// synced via a GitHub repo — it's small, diffable, and merge-friendly.
type State struct {
	Version    int      `json:"version"`
	Read       []string `json:"read"`
	LastSynced string   `json:"last_synced,omitempty"`
}

// LoadState reads the state file from disk. If the file doesn't exist,
// returns an empty state — this is not an error (first run).
func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &State{Version: 1, Read: []string{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading state file %s: %w", path, err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing state file %s: %w", path, err)
	}
	return &s, nil
}

// SaveState writes the state to disk using a write-to-temp-then-rename
// pattern for crash safety.
func SaveState(path string, state *State) error {
	return writeJSON(path, state)
}

// IsRead reports whether an article GUID has been marked as read.
func (s *State) IsRead(guid string) bool {
	for _, g := range s.Read {
		if g == guid {
			return true
		}
	}
	return false
}

// MarkRead adds a GUID to the read list if it isn't already there.
func (s *State) MarkRead(guid string) {
	if !s.IsRead(guid) {
		s.Read = append(s.Read, guid)
	}
}

// MarkUnread removes a GUID from the read list.
func (s *State) MarkUnread(guid string) {
	for i, g := range s.Read {
		if g == guid {
			// LEARN: This is the standard Go slice deletion pattern —
			// replace the element with the last one and shrink the slice.
			// Order doesn't matter for the read list.
			s.Read[i] = s.Read[len(s.Read)-1]
			s.Read = s.Read[:len(s.Read)-1]
			return
		}
	}
}

// writeJSON marshals a value to JSON and writes it to path atomically.
func writeJSON(path string, v any) error {
	// WHY: Write to a temp file then rename — prevents partial reads if
	// the process crashes mid-write. rename(2) is atomic on POSIX.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", path, err)
	}

	tmpPath := path + ".tmp"
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("writing temp file %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming %s to %s: %w", tmpPath, path, err)
	}
	return nil
}
