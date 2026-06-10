package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

	// Tracked file: stored in the db repo, committed, symlinked from the project
	dbFile := filepath.Join(env.DBDir, "myproject", "main", "TASK.md")
	env.CreateFile(dbFile, "# Task")
	run(t, env.DBDir, "git", "add", dbFile)
	run(t, env.DBDir, "git", "commit", "-m", "initial")

	testFile := filepath.Join(repoDir, "TASK.md")
	if err := os.Symlink(dbFile, testFile); err != nil {
		t.Fatal(err)
	}

	// Modify the stored copy, then re-add: stages the modification
	if err := os.WriteFile(dbFile, []byte("# Task v2"), 0644); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"add", "TASK.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add on tracked file failed: %v", err)
	}

	staged := gitOut(t, env.DBDir, "diff", "--cached", "--name-only")
	if !strings.Contains(staged, "myproject/main/TASK.md") {
		t.Errorf("modification should be staged, staged files: %q", staged)
	}
}

func TestAddCommand_NonexistentFile(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	env.InitDBRepo()

	var stderr bytes.Buffer
	rootCmd.SetErr(&stderr)
	defer rootCmd.SetErr(nil)

	rootCmd.SetArgs([]string{"add", "nope.md"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("add of nonexistent file should fail")
	}
	if strings.Contains(stderr.String(), "Usage:") {
		t.Error("runtime failure should not dump usage")
	}
}

func TestAddCommand_PartialFailure(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	env.InitDBRepo()

	env.CreateFile(filepath.Join(repoDir, "good.md"), "# Good")

	rootCmd.SetArgs([]string{"add", "good.md", "nope.md"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("partial failure should return an error")
	}

	if !env.IsSymlink(filepath.Join(repoDir, "good.md")) {
		t.Error("good.md should still be added despite the failing item")
	}
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
