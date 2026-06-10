package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KakkoiDev/aidb/internal/testutil"
)

func TestPushCommand_Basic(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	remoteDir := setupPullEnv(t, env)

	env.CreateFile(filepath.Join(env.DBDir, "local.txt"), "local")
	run(t, env.DBDir, "git", "add", ".")
	run(t, env.DBDir, "git", "commit", "-m", "local commit")

	rootCmd.SetArgs([]string{"push"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("push failed: %v", err)
	}

	if gitOut(t, remoteDir, "rev-parse", "HEAD") != gitOut(t, env.DBDir, "rev-parse", "HEAD") {
		t.Error("remote HEAD should match local HEAD after push")
	}
}

func TestPushCommand_DivergedRemote(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	remoteDir := setupPullEnv(t, env)

	// Local commit + a remote commit pushed from elsewhere = diverged
	env.CreateFile(filepath.Join(env.DBDir, "local.txt"), "local")
	run(t, env.DBDir, "git", "add", ".")
	run(t, env.DBDir, "git", "commit", "-m", "local commit")
	pushToRemote(t, env, remoteDir, "remote.txt", "remote")

	rootCmd.SetArgs([]string{"push"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("push should recover from divergence via rebase, got: %v", err)
	}

	remoteLog := gitOut(t, remoteDir, "log", "--oneline")
	if !strings.Contains(remoteLog, "local commit") {
		t.Errorf("remote should contain the local commit, log:\n%s", remoteLog)
	}
	if !strings.Contains(remoteLog, "add remote.txt") {
		t.Errorf("remote commit should be preserved, log:\n%s", remoteLog)
	}
}

func TestPushCommand_NoRemote(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	if err := os.MkdirAll(env.DBDir, 0755); err != nil {
		t.Fatal(err)
	}
	run(t, env.DBDir, "git", "init")

	rootCmd.SetArgs([]string{"push"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("push should fail when no remote configured")
	}
	if !strings.Contains(err.Error(), "no remote configured") {
		t.Errorf("expected 'no remote configured' error, got: %v", err)
	}
}
