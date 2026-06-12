package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func loadSyncState(storeDir string) (map[string]IndexEntry, error) {
	path := filepath.Join(storeDir, ".sync_state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]IndexEntry), nil
		}
		return nil, err
	}
	var state map[string]IndexEntry
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return state, nil
}

func saveSyncState(storeDir string, state map[string]IndexEntry) error {
	path := filepath.Join(storeDir, ".sync_state.json")
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
