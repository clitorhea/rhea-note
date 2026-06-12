package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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

	var index map[string]IndexEntry
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, err
	}
	return index, nil
}

func (c *Client) DownloadNote(noteID string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.ServerURL+"/notes/"+noteID, nil)
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
	req, err := http.NewRequest("PUT", c.ServerURL+"/notes/"+noteID, bytes.NewReader(payload))
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
	req, err := http.NewRequest("DELETE", c.ServerURL+"/notes/"+noteID, nil)
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
