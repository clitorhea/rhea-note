package storage

import (
	"os"
	"path/filepath"
	"time"
)

type LocalNoteInfo struct {
	ID        string
	UpdatedAt time.Time
}

// SaveNoteLocal writes the encrypted payload to a local file.
func SaveNoteLocal(storeDir, noteID string, payload []byte) error {
	if err := os.MkdirAll(storeDir, 0700); err != nil {
		return err
	}
	notePath := filepath.Join(storeDir, noteID+".enc")
	return os.WriteFile(notePath, payload, 0600)
}

// LoadNoteLocal reads the encrypted payload from a local file.
func LoadNoteLocal(storeDir, noteID string) ([]byte, error) {
	notePath := filepath.Join(storeDir, noteID+".enc")
	return os.ReadFile(notePath)
}

// ListNotesLocal returns a list of local note IDs.
func ListNotesLocal(storeDir string) ([]string, error) {
	entries, err := os.ReadDir(storeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var notes []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".enc" {
			name := entry.Name()
			notes = append(notes, name[:len(name)-4]) // strip .enc
		}
	}
	return notes, nil
}

// ListNotesLocalWithTime returns a map of local notes and their modification times.
func ListNotesLocalWithTime(storeDir string) (map[string]LocalNoteInfo, error) {
	entries, err := os.ReadDir(storeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]LocalNoteInfo{}, nil
		}
		return nil, err
	}

	notes := make(map[string]LocalNoteInfo)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".enc" {
			name := entry.Name()
			id := name[:len(name)-4]
			info, err := entry.Info()
			if err == nil {
				notes[id] = LocalNoteInfo{
					ID:        id,
					UpdatedAt: info.ModTime(),
				}
			}
		}
	}
	return notes, nil
}
