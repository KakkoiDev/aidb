package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/KakkoiDev/aidb/internal/testutil"
)

func TestInitCommand_Basic(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	// Run init command
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify directory exists
	if !env.FileExists(env.DBDir) {
		t.Error("~/.aidb should exist")
	}

	// Verify git initialized
	gitDir := filepath.Join(env.DBDir, ".git")
	if !env.FileExists(gitDir) {
		t.Error("~/.aidb/.git should exist")
	}
}

func TestInitCommand_WithRemote(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	remoteURL := "git@github.com:user/kb.git"

	// Run init with remote
	rootCmd.SetArgs([]string{"init", "--remote", remoteURL})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify remote configured
	cmd := exec.Command("git", "-C", env.DBDir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get remote: %v", err)
	}

	got := strings.TrimSpace(string(out))
	if got != remoteURL {
		t.Errorf("remote = %q, want %q", got, remoteURL)
	}
}

func TestInitCommand_UpdateExistingRemote(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	// Initialize with one remote
	env.InitDBRepo()
	exec.Command("git", "-C", env.DBDir, "remote", "add", "origin", "git@old.com:old/repo.git").Run()

	newURL := "git@github.com:user/new.git"

	// Run init with different remote
	rootCmd.SetArgs([]string{"init", "--remote", newURL})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify remote updated
	cmd := exec.Command("git", "-C", env.DBDir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get remote: %v", err)
	}

	got := strings.TrimSpace(string(out))
	if got != newURL {
		t.Errorf("remote = %q, want %q", got, newURL)
	}
}

func TestInitCommand_SameRemoteNoOp(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	remoteURL := "git@github.com:user/kb.git"

	// Initialize with remote
	env.InitDBRepo()
	exec.Command("git", "-C", env.DBDir, "remote", "add", "origin", remoteURL).Run()

	// Run init with same remote - should not error
	rootCmd.SetArgs([]string{"init", "--remote", remoteURL})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("init command failed: %v", err)
	}
}

func TestHasRemote(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	env.InitDBRepo()

	// No remote yet
	if HasRemote(env.DBDir) {
		t.Error("HasRemote should return false without remote")
	}

	// Add remote
	exec.Command("git", "-C", env.DBDir, "remote", "add", "origin", "git@github.com:user/kb.git").Run()

	if !HasRemote(env.DBDir) {
		t.Error("HasRemote should return true with remote")
	}
}

func TestGetRemoteURL(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	env.InitDBRepo()

	// No remote
	if url := GetRemoteURL(env.DBDir); url != "" {
		t.Errorf("GetRemoteURL = %q, want empty", url)
	}

	// With remote
	expected := "git@github.com:user/kb.git"
	exec.Command("git", "-C", env.DBDir, "remote", "add", "origin", expected).Run()

	if url := GetRemoteURL(env.DBDir); url != expected {
		t.Errorf("GetRemoteURL = %q, want %q", url, expected)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	// Init with initial commit
	repoDir := env.InitGitRepoWithBranch("testrepo", "main")

	branch := GetCurrentBranch(repoDir)
	// Branch could be "main" or "master" depending on git config
	if branch != "main" && branch != "master" {
		t.Errorf("GetCurrentBranch = %q, want main or master", branch)
	}
}

func TestIsInitialized(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	// Get real config (uses test HOME)
	cfg, err := newTestConfig(env)
	if err != nil {
		t.Fatal(err)
	}

	// Not initialized
	if IsInitialized(cfg) {
		t.Error("IsInitialized should return false before init")
	}

	// Initialize
	os.MkdirAll(env.DBDir, 0755)

	if !IsInitialized(cfg) {
		t.Error("IsInitialized should return true after init")
	}
}

// newTestConfig creates a config using test environment
func newTestConfig(env *testutil.TestEnv) (*config.Config, error) {
	return &config.Config{
		HomeDir: env.HomeDir,
		DBDir:   env.DBDir,
	}, nil
}
