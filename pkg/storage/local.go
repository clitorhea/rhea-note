package storage

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalNoteInfo struct {
	ID        string
	UpdatedAt time.Time
}

func EncodeID(id string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(id))
}

func DecodeID(encoded string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// SaveNoteLocal writes the encrypted payload to a local file.
func SaveNoteLocal(storeDir, noteID string, payload []byte) error {
	if err := os.MkdirAll(storeDir, 0700); err != nil {
		return err
	}
	notePath := filepath.Join(storeDir, EncodeID(noteID)+".enc")
	return os.WriteFile(notePath, payload, 0600)
}

// LoadNoteLocal reads the encrypted payload from a local file.
func LoadNoteLocal(storeDir, noteID string) ([]byte, error) {
	notePath := filepath.Join(storeDir, EncodeID(noteID)+".enc")
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
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".enc") {
			name := entry.Name()
			encodedID := name[:len(name)-4] // strip .enc
			id, err := DecodeID(encodedID)
			if err == nil {
				notes = append(notes, id)
			}
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
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".enc") {
			name := entry.Name()
			encodedID := name[:len(name)-4]
			id, err := DecodeID(encodedID)
			if err == nil {
				info, err := entry.Info()
				if err == nil {
					notes[id] = LocalNoteInfo{
						ID:        id,
						UpdatedAt: info.ModTime(),
					}
				}
			}
		}
	}
	return notes, nil
}
