package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Metadata stores file tracking information
type Metadata struct {
	Version int                  `json:"version"`
	Files   map[string]*FileInfo `json:"files"`
	path    string
}

// FileInfo stores per-file metadata
type FileInfo struct {
	Seen   bool      `json:"seen"`
	Hash   string    `json:"hash"`
	SeenAt time.Time `json:"seenAt,omitempty"`
}

// New creates or loads metadata from path
func New(dbDir string) (*Metadata, error) {
	path := filepath.Join(dbDir, ".metadata.json")
	m := &Metadata{
		Version: 1,
		Files:   make(map[string]*FileInfo),
		path:    path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, m); err != nil {
		return nil, err
	}
	m.path = path
	return m, nil
}

// Save writes metadata to disk
func (m *Metadata) Save() error {
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0644)
}

// MarkSeen marks a file as seen with current hash
func (m *Metadata) MarkSeen(relPath string, hash string) {
	m.Files[relPath] = &FileInfo{
		Seen:   true,
		Hash:   hash,
		SeenAt: time.Now().UTC(),
	}
}

// MarkUnseen marks a file as unseen
func (m *Metadata) MarkUnseen(relPath string) {
	if info, ok := m.Files[relPath]; ok {
		info.Seen = false
	}
}

// IsSeen returns true if file was seen and hash matches
func (m *Metadata) IsSeen(relPath string, currentHash string) bool {
	info, ok := m.Files[relPath]
	if !ok {
		return false
	}
	// If hash changed, mark as unseen
	if info.Hash != currentHash {
		info.Seen = false
		return false
	}
	return info.Seen
}

// GetInfo returns file info or nil
func (m *Metadata) GetInfo(relPath string) *FileInfo {
	return m.Files[relPath]
}

// Remove deletes file from metadata
func (m *Metadata) Remove(relPath string) {
	delete(m.Files, relPath)
}

// HashFile computes SHA256 hash of file content
func HashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(hash[:]), nil
}
