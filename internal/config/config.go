package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Config holds aidb configuration
type Config struct {
	HomeDir     string
	DBDir       string // ~/.claude-database (keeping original location for compatibility)
	ProjectName string
	TaskName    string
	ProjectDir  string // Full path: DBDir/ProjectName/TaskName
}

// New creates a new Config with defaults
func New() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &Config{
		HomeDir: homeDir,
		DBDir:   filepath.Join(homeDir, ".claude-database"),
	}, nil
}

// SetProject sets the current project context
func (c *Config) SetProject(project, task string) {
	c.ProjectName = project
	c.TaskName = task
	c.ProjectDir = filepath.Join(c.DBDir, project, task)
}

// GetProjectFromCwd detects project from current working directory
func (c *Config) GetProjectFromCwd() (project, task string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	// Try to detect git repo name
	project = getGitRepoName(cwd)
	if project == "" {
		// Fall back to directory name
		project = filepath.Base(cwd)
	}

	// Try to get git branch for task name
	task = getGitBranch(cwd)
	if task == "" {
		task = "main"
	}

	return project, task, nil
}

// EnsureProjectDir creates the project directory if it doesn't exist
func (c *Config) EnsureProjectDir() error {
	return os.MkdirAll(c.ProjectDir, 0755)
}

// GetFilePath returns the full path for a file in the project directory
func (c *Config) GetFilePath(filename string) string {
	return filepath.Join(c.ProjectDir, filename)
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
