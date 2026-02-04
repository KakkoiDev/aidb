package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KakkoiDev/aidb/internal/testutil"
)

func TestAddCommand_SingleFile(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	// Create a git repo and initialize aidb
	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	// Initialize aidb directory as git repo
	env.InitDBRepo()

	// Create a test file
	testFile := filepath.Join(repoDir, "TASK.md")
	if err := os.WriteFile(testFile, []byte("# Task"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run add command
	rootCmd.SetArgs([]string{"add", "TASK.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add command failed: %v", err)
	}

	// Verify symlink was created
	info, err := os.Lstat(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("TASK.md should be a symlink")
	}

	// Verify file exists in database (check symlink target)
	target, err := os.Readlink(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if !env.FileExists(target) {
		t.Errorf("symlink target should exist: %s", target)
	}

	// Verify content preserved
	content, _ := os.ReadFile(testFile)
	if string(content) != "# Task" {
		t.Errorf("content = %q, want %q", string(content), "# Task")
	}
}

func TestAddCommand_AlreadyTracked(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	env.InitDBRepo()

	// Create file directly in database
	dbFile := filepath.Join(env.DBDir, "myproject", "main", "TASK.md")
	if err := os.MkdirAll(filepath.Dir(dbFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dbFile, []byte("# Task"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create symlink pointing to it
	testFile := filepath.Join(repoDir, "TASK.md")
	if err := os.Symlink(dbFile, testFile); err != nil {
		t.Fatal(err)
	}

	// Try to add - should report already tracked
	rootCmd.SetArgs([]string{"add", "TASK.md"})
	// Note: command doesn't error, just prints message
	_ = rootCmd.Execute()
}

func TestAddCommand_GlobPattern(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	env.InitDBRepo()

	// Create multiple md files
	for _, name := range []string{"TASK.md", "MEMO.md", "README.md"} {
		if err := os.WriteFile(filepath.Join(repoDir, name), []byte("# "+name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Add all md files
	rootCmd.SetArgs([]string{"add", "*.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add command failed: %v", err)
	}

	// Verify all are symlinks
	for _, name := range []string{"TASK.md", "MEMO.md", "README.md"} {
		info, err := os.Lstat(filepath.Join(repoDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s should be a symlink", name)
		}
	}
}
