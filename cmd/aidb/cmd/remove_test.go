package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KakkoiDev/aidb/internal/testutil"
)

func TestRemoveCommand(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "feature")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	env.InitDBRepo()

	// Create file in database
	dbDir := filepath.Join(env.DBDir, "myproject", "feature")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatal(err)
	}
	dbFile := filepath.Join(dbDir, "TASK.md")
	if err := os.WriteFile(dbFile, []byte("# Task content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create symlink
	linkPath := filepath.Join(repoDir, "TASK.md")
	if err := os.Symlink(dbFile, linkPath); err != nil {
		t.Fatal(err)
	}

	// Run remove command
	rootCmd.SetArgs([]string{"remove", "TASK.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("remove command failed: %v", err)
	}

	// Verify it's no longer a symlink
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("TASK.md should not be a symlink after remove")
	}

	// Verify content preserved
	content, _ := os.ReadFile(linkPath)
	if string(content) != "# Task content" {
		t.Errorf("content = %q, want %q", string(content), "# Task content")
	}

	// Verify file removed from database
	if env.FileExists(dbFile) {
		t.Error("file should be removed from database")
	}
}

func TestRemoveCommand_NotTracked(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepo("myproject")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	// Create regular file (not tracked)
	if err := os.WriteFile(filepath.Join(repoDir, "test.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Try to remove - should error
	rootCmd.SetArgs([]string{"remove", "test.md"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for non-tracked file")
	}
}
