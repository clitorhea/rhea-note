package sync

import (
	"fmt"
	"time"

	"github.com/clitorhea/rhea-note/pkg/logger"
	"github.com/clitorhea/rhea-note/pkg/storage"
)

func (c *Client) Synchronize(storeDir string) error {
	remoteIndex, err := c.FetchIndex()
	if err != nil {
		return fmt.Errorf("failed to fetch remote index: %v", err)
	}

	localNotes, err := storage.ListNotesLocalWithTime(storeDir)
	if err != nil {
		return fmt.Errorf("failed to list local notes: %v", err)
	}

	lastSyncState, err := loadSyncState(storeDir)
	if err != nil {
		return fmt.Errorf("failed to load sync state: %v", err)
	}

	// 1. Process local notes
	for id, localNote := range localNotes {
		remoteNote, existsRemote := remoteIndex[id]
		lastSyncNote, existsLastSync := lastSyncState[id]

		if !existsRemote {
			if existsLastSync {
				// Was previously synced, but now missing remotely -> DELETED remotely!
				logger.Infof("Deleting local note because it was deleted remotely: %s", id)
				storage.DeleteNoteLocal(storeDir, id)
				continue
			}

			// Doesn't exist on remote and no sync state -> completely new. Push it.
			logger.Infof("Pushing new local note: %s", id)
			payload, err := storage.LoadNoteLocal(storeDir, id)
			if err == nil {
				c.UploadNote(id, payload)
			}
			continue
		}

		// Conflict Guard Logic
		localChanged := !existsLastSync || localNote.UpdatedAt.After(lastSyncNote.UpdatedAt.Add(time.Second))
		remoteChanged := !existsLastSync || remoteNote.UpdatedAt.After(lastSyncNote.UpdatedAt.Add(time.Second))

		if localChanged && remoteChanged {
			// Check if contents are actually identical before forking
			localPayload, _ := storage.LoadNoteLocal(storeDir, id)
			remotePayload, err := c.DownloadNote(id)
			if err == nil && string(localPayload) == string(remotePayload) {
				logger.Infof("Files identical, skipping conflict for: %s", id)
				continue
			}

			// Both changed and contents differ! Create a fork.
			conflictID := fmt.Sprintf("%s_conflict_%s", id, time.Now().Format("2006-01-02-150405"))
			logger.Infof("Conflict detected for %s. Forking local to %s", id, conflictID)
			
			// Rename local to conflict ID
			storage.SaveNoteLocal(storeDir, conflictID, localPayload)
			
			// Remote payload already downloaded, just save it locally
			if err == nil {
				storage.SaveNoteLocal(storeDir, id, remotePayload)
			}
			
			// Push the fork to the remote
			c.UploadNote(conflictID, localPayload)
			
		} else if localChanged && !remoteChanged {
			// Only local changed
			logger.Infof("Pushing updated local note: %s", id)
			payload, err := storage.LoadNoteLocal(storeDir, id)
			if err == nil {
				c.UploadNote(id, payload)
			}
		} else if !localChanged && remoteChanged {
			// Only remote changed
			logger.Infof("Pulling updated remote note: %s", id)
			payload, err := c.DownloadNote(id)
			if err == nil {
				storage.SaveNoteLocal(storeDir, id, payload)
			}
		}
	}

	// 2. Process remote notes that are not local
	for id := range remoteIndex {
		if _, exists := localNotes[id]; !exists {
			if _, existsLastSync := lastSyncState[id]; existsLastSync {
				// Was previously synced, but now missing locally -> DELETED locally!
				logger.Infof("Deleting remote note because it was deleted locally: %s", id)
				c.DeleteNote(id)
			} else {
				logger.Infof("Pulling new remote note: %s", id)
				payload, err := c.DownloadNote(id)
				if err == nil {
					storage.SaveNoteLocal(storeDir, id, payload)
				}
			}
		}
	}

	// Refetch the remote index after all operations to save as the new state
	finalIndex, err := c.FetchIndex()
	if err == nil {
		saveSyncState(storeDir, finalIndex)
	}

	return nil
}
