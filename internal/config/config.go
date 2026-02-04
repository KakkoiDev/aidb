package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Config holds aidb configuration
type Config struct {
	HomeDir string
	DBDir   string // ~/.aidb
}

// New creates a new Config with defaults
func New() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &Config{
		HomeDir: homeDir,
		DBDir:   filepath.Join(homeDir, ".aidb"),
	}, nil
}

// IsGitRepo returns true if current directory is inside a git repository
func (c *Config) IsGitRepo() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	return getGitRepoName(cwd) != ""
}

// GetProjectFromCwd detects project/branch from current working directory
func (c *Config) GetProjectFromCwd() (project, branch string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	// Try to detect git repo name
	project = getGitRepoName(cwd)
	if project != "" {
		// Git repo: use repo name and branch
		branch = getGitBranch(cwd)
		if branch == "" {
			branch = "main"
		}
		return project, branch, nil
	}

	// Non-git: use directory name, default branch
	project = filepath.Base(cwd)
	return project, "main", nil
}

// GetStoragePath returns the storage path for a file based on current context
// Git repos: ~/.aidb/{repo-name}/{branch}/{filename}
// Non-git: ~/.aidb/{path-from-home}/{filename}
func (c *Config) GetStoragePath(filename string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	repoName := getGitRepoName(cwd)
	if repoName != "" {
		// Git repo: use repo name and branch
		branch := getGitBranch(cwd)
		if branch == "" {
			branch = "main"
		}
		return filepath.Join(c.DBDir, repoName, branch, filename), nil
	}

	// Non-git: use path relative to home
	// Resolve symlinks for consistent path comparison
	homeResolved, _ := filepath.EvalSymlinks(c.HomeDir)
	cwdResolved, _ := filepath.EvalSymlinks(cwd)
	if homeResolved == "" {
		homeResolved = c.HomeDir
	}
	if cwdResolved == "" {
		cwdResolved = cwd
	}

	relPath, err := filepath.Rel(homeResolved, cwdResolved)
	if err != nil {
		// Fallback to directory name if can't get relative path
		relPath = filepath.Base(cwd)
	}
	return filepath.Join(c.DBDir, relPath, filename), nil
}

// EnsureStorageDir creates and returns the storage directory for current context
func (c *Config) EnsureStorageDir() (string, error) {
	path, err := c.GetStoragePath("dummy")
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// EnsureDBDir creates the base database directory and initializes git if needed
func (c *Config) EnsureDBDir() error {
	if err := os.MkdirAll(c.DBDir, 0755); err != nil {
		return err
	}

	// Check if git is initialized
	gitDir := filepath.Join(c.DBDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Initialize git repo
		cmd := exec.Command("git", "-C", c.DBDir, "init")
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

// getGitRepoName returns the git repository name
func getGitRepoName(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return filepath.Base(strings.TrimSpace(string(out)))
}

// getGitBranch returns the current git branch
func getGitBranch(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
