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

func TestAddCommand_AbsolutePath(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	env.InitDBRepo()

	testFile := filepath.Join(repoDir, "TASK.md")
	env.CreateFile(testFile, "# Task")

	rootCmd.SetArgs([]string{"add", testFile})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add with absolute path failed: %v", err)
	}
	if !env.IsSymlink(testFile) {
		t.Error("TASK.md should be a symlink after absolute-path add")
	}
}

func TestAddCommand_SubdirSameStorePath(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	env.InitDBRepo()

	env.CreateFile(filepath.Join(repoDir, "docs", "x.md"), "# X")
	env.CreateFile(filepath.Join(repoDir, "docs", "y.md"), "# Y")

	// Add one file from the repo root, the other from inside docs/
	rootCmd.SetArgs([]string{"add", "docs/x.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add from root failed: %v", err)
	}
	if err := os.Chdir(filepath.Join(repoDir, "docs")); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"add", "y.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add from subdir failed: %v", err)
	}

	relX, err := filepath.Rel(env.DBDir, env.SymlinkTarget(filepath.Join(repoDir, "docs", "x.md")))
	if err != nil {
		t.Fatal(err)
	}
	relY, err := filepath.Rel(env.DBDir, env.SymlinkTarget(filepath.Join(repoDir, "docs", "y.md")))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(relX) != filepath.Dir(relY) {
		t.Errorf("store layout depends on cwd: %q vs %q", relX, relY)
	}
}

func TestAddCommand_ParentEscapeRefused(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	env.InitDBRepo()

	outside := filepath.Join(env.WorkDir, "outside.md")
	env.CreateFile(outside, "# Outside")

	rootCmd.SetArgs([]string{"add", "../outside.md"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("add escaping the project should be refused")
	}
	if env.IsSymlink(outside) {
		t.Error("outside.md must not be moved into the store")
	}
	if !env.FileExists(outside) {
		t.Error("outside.md must remain in place")
	}
}

func TestAddCommand_SameBasenameCollisionRefused(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	env.InitDBRepo()

	// Two distinct projects sharing the directory basename "proj"
	repoA := env.InitGitRepoWithBranch(filepath.Join("a", "proj"), "main")
	repoB := env.InitGitRepoWithBranch(filepath.Join("b", "proj"), "main")

	if err := os.Chdir(repoA); err != nil {
		t.Fatal(err)
	}
	env.CreateFile(filepath.Join(repoA, "TASK.md"), "# A")
	rootCmd.SetArgs([]string{"add", "TASK.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	if err := os.Chdir(repoB); err != nil {
		t.Fatal(err)
	}
	env.CreateFile(filepath.Join(repoB, "NOTES.md"), "# B")
	rootCmd.SetArgs([]string{"add", "NOTES.md"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("add into a namespace pinned to another repo should be refused")
	}
	if !strings.Contains(err.Error(), "a/proj") || !strings.Contains(err.Error(), "b/proj") {
		t.Errorf("error should name both repo paths, got: %v", err)
	}
	if env.IsSymlink(filepath.Join(repoB, "NOTES.md")) {
		t.Error("refused add must not move the file")
	}
}

func TestAddCommand_SameRepoSecondAddOK(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	env.InitDBRepo()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	env.CreateFile(filepath.Join(repoDir, "x.md"), "# X")
	rootCmd.SetArgs([]string{"add", "x.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	env.CreateFile(filepath.Join(repoDir, "y.md"), "# Y")
	rootCmd.SetArgs([]string{"add", "y.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("second add from the same repo failed: %v", err)
	}
}

func TestAddCommand_LazyPinExistingNamespace(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	env.InitDBRepo()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	branch := gitOut(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")

	// Namespace predates origin pinning: files exist, no .origin
	env.CreateFile(filepath.Join(env.DBDir, "myproject", branch, "old.md"), "# Old")

	env.CreateFile(filepath.Join(repoDir, "new.md"), "# New")
	rootCmd.SetArgs([]string{"add", "new.md"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add into pre-pin namespace failed: %v", err)
	}

	if !env.FileExists(filepath.Join(env.DBDir, "myproject", ".origin")) {
		t.Error("add should lazily pin the namespace with .origin")
	}
}

func TestAddCommand_DetachedHeadRefused(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	env.InitDBRepo()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	run(t, repoDir, "git", "checkout", "--detach")

	env.CreateFile(filepath.Join(repoDir, "TASK.md"), "# Task")
	rootCmd.SetArgs([]string{"add", "TASK.md"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("add from a detached HEAD should be refused")
	}
	if !strings.Contains(err.Error(), "detached HEAD") {
		t.Errorf("expected detached HEAD error, got: %v", err)
	}
	if env.IsSymlink(filepath.Join(repoDir, "TASK.md")) {
		t.Error("refused add must not move the file")
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
