package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/KakkoiDev/aidb/internal/testutil"
)

func TestBackupRun_NoRemote(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	env.InitDBRepo()

	env.CreateFile(filepath.Join(env.DBDir, "note.md"), "content")

	rootCmd.SetArgs([]string{"backup-run"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("backup-run without remote should commit locally and exit 0, got: %v", err)
	}

	log := gitOut(t, env.DBDir, "log", "--oneline")
	if !strings.Contains(log, "Auto-backup") {
		t.Errorf("expected an Auto-backup commit, log:\n%s", log)
	}
}

func TestBackupRun_DivergedRemote(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	remoteDir := setupPullEnv(t, env)

	// Remote moved ahead while this machine accumulated uncommitted changes
	pushToRemote(t, env, remoteDir, "remote.txt", "remote")
	env.CreateFile(filepath.Join(env.DBDir, "note.md"), "content")

	rootCmd.SetArgs([]string{"backup-run"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("backup-run should recover from divergence via rebase, got: %v", err)
	}

	remoteLog := gitOut(t, remoteDir, "log", "--oneline")
	if !strings.Contains(remoteLog, "Auto-backup") {
		t.Errorf("remote should contain the backup commit, log:\n%s", remoteLog)
	}
}

func TestBackupRun_NoChanges(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()
	env.InitDBRepo()

	rootCmd.SetArgs([]string{"backup-run"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("backup-run on clean tree should exit 0, got: %v", err)
	}
}
