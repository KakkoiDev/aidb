package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KakkoiDev/aidb/internal/testutil"
)

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

// addAndCommit tracks a file via aidb add + commit and returns the stored path
func addAndCommit(t *testing.T, env *testutil.TestEnv, repoDir, name, content string) string {
	t.Helper()

	env.CreateFile(filepath.Join(repoDir, name), content)
	rootCmd.SetArgs([]string{"add", name})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	rootCmd.SetArgs([]string{"commit", "initial " + name})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	target, err := os.Readlink(filepath.Join(repoDir, name))
	if err != nil {
		t.Fatal(err)
	}
	return target
}

func TestCommitCommand_TrackedModification(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	env.InitDBRepo()

	stored := addAndCommit(t, env, repoDir, "TASK.md", "# Task")

	if err := os.WriteFile(stored, []byte("# Task v2"), 0644); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"commit", "update"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	relPath, err := filepath.Rel(env.DBDir, stored)
	if err != nil {
		t.Fatal(err)
	}
	got := gitOut(t, env.DBDir, "show", "HEAD:"+relPath)
	if got != "# Task v2" {
		t.Errorf("HEAD content = %q, want %q", got, "# Task v2")
	}
	if status := gitOut(t, env.DBDir, "status", "--porcelain"); status != "" {
		t.Errorf("working tree not clean after commit:\n%s", status)
	}
}

func TestCommitCommand_TrackedDeletion(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	env.InitDBRepo()

	stored := addAndCommit(t, env, repoDir, "TASK.md", "# Task")

	if err := os.Remove(stored); err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{"commit", "delete"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	relPath, err := filepath.Rel(env.DBDir, stored)
	if err != nil {
		t.Fatal(err)
	}
	if files := gitOut(t, env.DBDir, "ls-files", "--", relPath); files != "" {
		t.Errorf("deleted file still tracked: %q", files)
	}
	if status := gitOut(t, env.DBDir, "status", "--porcelain"); status != "" {
		t.Errorf("working tree not clean after commit:\n%s", status)
	}
}

func TestCommitCommand_CleanTree(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "main")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}
	env.InitDBRepo()

	addAndCommit(t, env, repoDir, "TASK.md", "# Task")
	head := gitOut(t, env.DBDir, "rev-parse", "HEAD")

	rootCmd.SetArgs([]string{"commit", "noop"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("commit on clean tree should exit 0, got: %v", err)
	}
	if gitOut(t, env.DBDir, "rev-parse", "HEAD") != head {
		t.Error("commit on clean tree should not create a commit")
	}
}
