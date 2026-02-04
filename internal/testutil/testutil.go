package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestEnv provides isolated test environment
type TestEnv struct {
	T       *testing.T
	TempDir string
	HomeDir string
	WorkDir string
	DBDir   string
	OldHome string
	OldCwd  string
}

// New creates a new test environment
func New(t *testing.T) *TestEnv {
	t.Helper()

	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	workDir := filepath.Join(tempDir, "work")
	dbDir := filepath.Join(homeDir, ".aidb")

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}

	oldHome := os.Getenv("HOME")
	oldCwd, _ := os.Getwd()

	os.Setenv("HOME", homeDir)
	if err := os.Chdir(workDir); err != nil {
		t.Fatal(err)
	}

	return &TestEnv{
		T:       t,
		TempDir: tempDir,
		HomeDir: homeDir,
		WorkDir: workDir,
		DBDir:   dbDir,
		OldHome: oldHome,
		OldCwd:  oldCwd,
	}
}

// Cleanup restores original environment
func (e *TestEnv) Cleanup() {
	os.Setenv("HOME", e.OldHome)
	os.Chdir(e.OldCwd)
}

// InitGitRepo initializes a git repo in the work directory
func (e *TestEnv) InitGitRepo(name string) string {
	e.T.Helper()

	repoDir := filepath.Join(e.WorkDir, name)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		e.T.Fatal(err)
	}

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoDir
		if err := cmd.Run(); err != nil {
			e.T.Fatalf("failed to run %v: %v", args, err)
		}
	}

	return repoDir
}

// InitGitRepoWithBranch initializes a git repo with a specific branch
func (e *TestEnv) InitGitRepoWithBranch(name, branch string) string {
	e.T.Helper()

	repoDir := e.InitGitRepo(name)

	// Create initial commit so we can create branches
	dummyFile := filepath.Join(repoDir, ".gitkeep")
	if err := os.WriteFile(dummyFile, []byte{}, 0644); err != nil {
		e.T.Fatal(err)
	}

	cmds := [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	}

	if branch != "master" && branch != "main" {
		cmds = append(cmds, []string{"git", "checkout", "-b", branch})
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repoDir
		if err := cmd.Run(); err != nil {
			e.T.Fatalf("failed to run %v: %v", args, err)
		}
	}

	return repoDir
}

// CreateFile creates a file with content
func (e *TestEnv) CreateFile(path, content string) {
	e.T.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		e.T.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		e.T.Fatal(err)
	}
}

// ReadFile reads file content
func (e *TestEnv) ReadFile(path string) string {
	e.T.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		e.T.Fatal(err)
	}
	return string(data)
}

// FileExists checks if file exists
func (e *TestEnv) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsSymlink checks if path is a symlink
func (e *TestEnv) IsSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// SymlinkTarget returns the target of a symlink
func (e *TestEnv) SymlinkTarget(path string) string {
	e.T.Helper()

	target, err := os.Readlink(path)
	if err != nil {
		e.T.Fatal(err)
	}
	return target
}

// InitDBRepo initializes the aidb repository
func (e *TestEnv) InitDBRepo() {
	e.T.Helper()

	if err := os.MkdirAll(e.DBDir, 0755); err != nil {
		e.T.Fatal(err)
	}

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = e.DBDir
		if err := cmd.Run(); err != nil {
			e.T.Fatalf("failed to run %v: %v", args, err)
		}
	}
}
