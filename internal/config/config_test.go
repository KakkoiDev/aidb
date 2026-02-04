package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KakkoiDev/aidb/internal/testutil"
)

func TestNew(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	cfg, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HomeDir != env.HomeDir {
		t.Errorf("HomeDir = %q, want %q", cfg.HomeDir, env.HomeDir)
	}

	expectedDBDir := filepath.Join(env.HomeDir, ".aidb")
	if cfg.DBDir != expectedDBDir {
		t.Errorf("DBDir = %q, want %q", cfg.DBDir, expectedDBDir)
	}
}

func TestGetProjectFromCwd_GitRepo(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "feature-x")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	cfg, _ := New()
	project, branch, err := cfg.GetProjectFromCwd()
	if err != nil {
		t.Fatal(err)
	}

	if project != "myproject" {
		t.Errorf("project = %q, want %q", project, "myproject")
	}
	if branch != "feature-x" {
		t.Errorf("branch = %q, want %q", branch, "feature-x")
	}
}

func TestGetProjectFromCwd_NonGit(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	nonGitDir := filepath.Join(env.WorkDir, "Documents", "notes")
	if err := os.MkdirAll(nonGitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(nonGitDir); err != nil {
		t.Fatal(err)
	}

	cfg, _ := New()
	project, branch, err := cfg.GetProjectFromCwd()
	if err != nil {
		t.Fatal(err)
	}

	// Non-git returns directory name
	if project != "notes" {
		t.Errorf("project = %q, want %q", project, "notes")
	}
	// Non-git defaults to "main"
	if branch != "main" {
		t.Errorf("branch = %q, want %q", branch, "main")
	}
}

func TestGetStoragePath_GitRepo(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "feature-x")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	cfg, _ := New()
	path, err := cfg.GetStoragePath("TASK.md")
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(env.HomeDir, ".aidb", "myproject", "feature-x", "TASK.md")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
}

func TestGetStoragePath_NonGit(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	// Create directory inside home (not work)
	nonGitDir := filepath.Join(env.HomeDir, "Documents", "notes")
	if err := os.MkdirAll(nonGitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(nonGitDir); err != nil {
		t.Fatal(err)
	}

	cfg, _ := New()
	path, err := cfg.GetStoragePath("ideas.md")
	if err != nil {
		t.Fatal(err)
	}

	// Non-git uses path from home
	expected := filepath.Join(env.HomeDir, ".aidb", "Documents", "notes", "ideas.md")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
}

func TestIsGitRepo(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	// Test git repo
	repoDir := env.InitGitRepo("testrepo")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	cfg, _ := New()
	if !cfg.IsGitRepo() {
		t.Error("IsGitRepo() = false, want true")
	}

	// Test non-git dir
	nonGitDir := filepath.Join(env.WorkDir, "notgit")
	if err := os.MkdirAll(nonGitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(nonGitDir); err != nil {
		t.Fatal(err)
	}

	if cfg.IsGitRepo() {
		t.Error("IsGitRepo() = true, want false")
	}
}

func TestEnsureStorageDir(t *testing.T) {
	env := testutil.New(t)
	defer env.Cleanup()

	repoDir := env.InitGitRepoWithBranch("myproject", "feature-x")
	if err := os.Chdir(repoDir); err != nil {
		t.Fatal(err)
	}

	cfg, _ := New()
	dir, err := cfg.EnsureStorageDir()
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(env.HomeDir, ".aidb", "myproject", "feature-x")
	if dir != expected {
		t.Errorf("dir = %q, want %q", dir, expected)
	}

	if !env.FileExists(dir) {
		t.Errorf("directory %q was not created", dir)
	}
}
