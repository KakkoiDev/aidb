package metadata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_NoFile(t *testing.T) {
	tmpDir := t.TempDir()

	m, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if m.Version != 1 {
		t.Errorf("Version = %d, want 1", m.Version)
	}
	if len(m.Files) != 0 {
		t.Errorf("Files = %v, want empty", m.Files)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	m, _ := New(tmpDir)
	m.MarkSeen("test/file.md", "sha256:abc123")
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	// Load fresh
	m2, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	info := m2.GetInfo("test/file.md")
	if info == nil {
		t.Fatal("info is nil")
	}
	if !info.Seen {
		t.Error("Seen = false, want true")
	}
	if info.Hash != "sha256:abc123" {
		t.Errorf("Hash = %q, want %q", info.Hash, "sha256:abc123")
	}
}

func TestIsSeen_HashChanged(t *testing.T) {
	tmpDir := t.TempDir()

	m, _ := New(tmpDir)
	m.MarkSeen("file.md", "sha256:original")

	// Same hash - still seen
	if !m.IsSeen("file.md", "sha256:original") {
		t.Error("IsSeen should be true for same hash")
	}

	// Different hash - becomes unseen
	if m.IsSeen("file.md", "sha256:changed") {
		t.Error("IsSeen should be false for changed hash")
	}

	// Verify it was marked unseen
	info := m.GetInfo("file.md")
	if info.Seen {
		t.Error("Seen should be false after hash change")
	}
}

func TestMarkUnseen(t *testing.T) {
	tmpDir := t.TempDir()

	m, _ := New(tmpDir)
	m.MarkSeen("file.md", "sha256:abc")
	m.MarkUnseen("file.md")

	if m.IsSeen("file.md", "sha256:abc") {
		t.Error("IsSeen should be false after MarkUnseen")
	}
}

func TestRemove(t *testing.T) {
	tmpDir := t.TempDir()

	m, _ := New(tmpDir)
	m.MarkSeen("file.md", "sha256:abc")
	m.Remove("file.md")

	if m.GetInfo("file.md") != nil {
		t.Error("GetInfo should return nil after Remove")
	}
}

func TestHashFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	// SHA256 of "hello world"
	expected := "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("hash = %q, want %q", hash, expected)
	}
}
