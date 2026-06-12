package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/clitorhea/rhea-note/pkg/storage"
)

type IndexEntry struct {
	NoteID    string    `json:"note_id"`
	UpdatedAt time.Time `json:"updated_at"`
	Hash      string    `json:"hash"`
}

type Client struct {
	ServerURL  string
	Token      string
	HTTPClient *http.Client
}

func NewClient(serverURL, token string) *Client {
	return &Client{
		ServerURL:  serverURL,
		Token:      token,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) FetchIndex() (map[string]IndexEntry, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/sync/index", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rawIndex map[string]IndexEntry
	if err := json.NewDecoder(resp.Body).Decode(&rawIndex); err != nil {
		return nil, err
	}
	
	decodedIndex := make(map[string]IndexEntry)
	for encodedID, entry := range rawIndex {
		id, err := storage.DecodeID(encodedID)
		if err == nil {
			entry.NoteID = id
			decodedIndex[id] = entry
		}
	}

	return decodedIndex, nil
}

func (c *Client) DownloadNote(noteID string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/notes/"+storage.EncodeID(noteID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) UploadNote(noteID string, payload []byte) error {
	req, err := http.NewRequest("PUT", c.ServerURL+"/notes/"+storage.EncodeID(noteID), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) DeleteNote(noteID string) error {
	req, err := http.NewRequest("DELETE", c.ServerURL+"/notes/"+storage.EncodeID(noteID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
