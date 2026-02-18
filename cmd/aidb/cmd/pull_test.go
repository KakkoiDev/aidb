package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KakkoiDev/aidb/internal/testutil"
)

// setupPullEnv creates a bare remote, clones it as ~/.aidb, and returns the remote path
func setupPullEnv(t *testing.T, env *testutil.TestEnv) string {
	t.Helper()

	// Create bare remote repo
	remoteDir := filepath.Join(env.TempDir, "remote.git")
	if err := os.MkdirAll(remoteDir, 0755); err != nil {
		t.Fatal(err)
	}
	run(t, remoteDir, "git", "init", "--bare")

	// Create a temp working copy to push initial commit
	tmpWork := filepath.Join(env.TempDir, "tmpwork")
	run(t, "", "git", "clone", remoteDir, tmpWork)
	run(t, tmpWork, "git", "config", "user.email", "test@test.com")
	run(t, tmpWork, "git", "config", "user.name", "Test")
	env.CreateFile(filepath.Join(tmpWork, "init.txt"), "initial")
	run(t, tmpWork, "git", "add", ".")
	run(t, tmpWork, "git", "commit", "-m", "initial")
	run(t, tmpWork, "git", "push")

	// Clone remote as ~/.aidb
	run(t, "", "git", "clone", remoteDir, env.DBDir)
	run(t, env.DBDir, "git", "config", "user.email", "test@test.com")
	run(t, env.DBDir, "git", "config", "user.name", "Test")

	return remoteDir
}

// pushToRemote creates a commit on the remote via a temp working copy
func pushToRemote(t *testing.T, env *testutil.TestEnv, remoteDir, filename, content string) {
	t.Helper()

	tmpWork := filepath.Join(env.TempDir, "tmpwork2")
	os.RemoveAll(tmpWork)
	run(t, "", "git", "clone", remoteDir, tmpWork)
	run(t, tmpWork, "git", "config", "user.email", "test@test.com")
	run(t, tmpWork, "git", "config", "user.name", "Test")
	env.CreateFile(filepath.Join(tmpWork, filename), content)
	run(t, tmpWork, "git", "add", ".")
	run(t, tmpWork, "git", "commit", "-m", "add "+filename)
	run(t, tmpWork, "git", "push")
}

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run %v: %v\n%s", args, err, out)
	}
}

func TestPullCommand_Basic(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	setupPullEnv(t, env)

	rootCmd.SetArgs([]string{"pull"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("pull command failed: %v", err)
	}
}

func TestPullCommand_Rebase(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	remoteDir := setupPullEnv(t, env)

	// Create local commit
	env.CreateFile(filepath.Join(env.DBDir, "local.txt"), "local")
	run(t, env.DBDir, "git", "add", ".")
	run(t, env.DBDir, "git", "commit", "-m", "local commit")

	// Push a different file to remote
	pushToRemote(t, env, remoteDir, "remote.txt", "remote")

	// Pull should rebase (no merge commit)
	rootCmd.SetArgs([]string{"pull"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("pull command failed: %v", err)
	}

	// Verify both files exist
	if !env.FileExists(filepath.Join(env.DBDir, "local.txt")) {
		t.Error("local.txt should exist after pull")
	}
	if !env.FileExists(filepath.Join(env.DBDir, "remote.txt")) {
		t.Error("remote.txt should exist after pull")
	}

	// Verify linear history (no merge commits)
	cmd := exec.Command("git", "-C", env.DBDir, "log", "--oneline", "--merges")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("expected no merge commits, got: %s", out)
	}
}

func TestPullCommand_StashUnstagedChanges(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	remoteDir := setupPullEnv(t, env)

	// Push a change to remote
	pushToRemote(t, env, remoteDir, "remote.txt", "remote")

	// Create unstaged local change
	env.CreateFile(filepath.Join(env.DBDir, "unstaged.txt"), "dirty")

	// Pull should succeed despite unstaged changes
	rootCmd.SetArgs([]string{"pull"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("pull command failed: %v", err)
	}

	// Verify unstaged file survives
	if !env.FileExists(filepath.Join(env.DBDir, "unstaged.txt")) {
		t.Error("unstaged.txt should survive pull")
	}
	got := env.ReadFile(filepath.Join(env.DBDir, "unstaged.txt"))
	if got != "dirty" {
		t.Errorf("unstaged.txt content = %q, want %q", got, "dirty")
	}

	// Verify remote change pulled
	if !env.FileExists(filepath.Join(env.DBDir, "remote.txt")) {
		t.Error("remote.txt should exist after pull")
	}
}

func TestPullCommand_NotInitialized(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	rootCmd.SetArgs([]string{"pull"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("pull should fail when not initialized")
	}
}
