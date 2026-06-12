package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type IndexEntry struct {
	NoteID    string    `json:"note_id"`
	UpdatedAt time.Time `json:"updated_at"`
	Hash      string    `json:"hash"`
}

type Server struct {
	StoreDir string
	Token    string
	index    map[string]IndexEntry
	mu       sync.RWMutex
}

func NewServer(storeDir, token string) *Server {
	os.MkdirAll(storeDir, 0700)
	
	s := &Server{
		StoreDir: storeDir,
		Token:    token,
		index:    make(map[string]IndexEntry),
	}
	
	// Rebuild index from existing files on disk
	entries, err := os.ReadDir(storeDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				info, err := entry.Info()
				if err == nil {
					s.index[entry.Name()] = IndexEntry{
						NoteID:    entry.Name(),
						UpdatedAt: info.ModTime(),
					}
				}
			}
		}
	}
	
	return s
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+s.Token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/sync/index") && r.Method == "GET" {
		s.authMiddleware(s.handleIndex)(w, r)
		return
	}
	
	if strings.HasPrefix(r.URL.Path, "/notes/") {
		noteID := strings.TrimPrefix(r.URL.Path, "/notes/")
		if noteID == "" {
			http.Error(w, "missing note id", http.StatusBadRequest)
			return
		}
		
		switch r.Method {
		case "GET":
			s.authMiddleware(func(w http.ResponseWriter, r *http.Request) { s.handleGetNote(w, r, noteID) })(w, r)
		case "PUT":
			s.authMiddleware(func(w http.ResponseWriter, r *http.Request) { s.handlePutNote(w, r, noteID) })(w, r)
		case "DELETE":
			s.authMiddleware(func(w http.ResponseWriter, r *http.Request) { s.handleDeleteNote(w, r, noteID) })(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	
	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.index)
}

func (s *Server) handleGetNote(w http.ResponseWriter, r *http.Request, noteID string) {
	path := filepath.Join(s.StoreDir, noteID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (s *Server) handlePutNote(w http.ResponseWriter, r *http.Request, noteID string) {
	// Enforce a strict 10MB limit on note size to prevent OOM
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "payload too large or read error", http.StatusRequestEntityTooLarge)
		return
	}
	
	path := filepath.Join(s.StoreDir, noteID)
	if err := os.WriteFile(path, data, 0600); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	info, err := os.Stat(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	s.mu.Lock()
	s.index[noteID] = IndexEntry{
		NoteID:    noteID,
		UpdatedAt: info.ModTime(),
	}
	s.mu.Unlock()
	
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeleteNote(w http.ResponseWriter, r *http.Request, noteID string) {
	path := filepath.Join(s.StoreDir, noteID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	s.mu.Lock()
	delete(s.index, noteID)
	s.mu.Unlock()
	
	w.WriteHeader(http.StatusNoContent)
}
