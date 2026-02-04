package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var initRemote string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize aidb database",
	Long: `Initialize the aidb database at ~/.aidb.

Creates the directory, initializes git, and optionally configures a remote.

Examples:
  aidb init                                    # Initialize ~/.aidb
  aidb init --remote git@github.com:user/kb.git  # With remote`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&initRemote, "remote", "", "Git remote URL to configure")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	// Create directory and init git
	if err := cfg.EnsureDBDir(); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Rename default branch to main
	gitCmd := exec.Command("git", "-C", cfg.DBDir, "branch", "-M", "main")
	gitCmd.Run() // Ignore error if branch doesn't exist yet

	printSuccess(fmt.Sprintf("Initialized %s", cfg.DBDir))

	// Configure remote if provided
	if initRemote != "" {
		if err := configureRemote(cfg.DBDir, initRemote); err != nil {
			return err
		}
		printSuccess(fmt.Sprintf("Remote configured: %s", initRemote))
	}

	return nil
}

func configureRemote(dir, url string) error {
	// Check if remote already exists
	checkCmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	out, err := checkCmd.Output()
	if err == nil {
		existingURL := strings.TrimSpace(string(out))
		if existingURL == url {
			return nil // Same URL, nothing to do
		}
		// Different URL, update it
		setCmd := exec.Command("git", "-C", dir, "remote", "set-url", "origin", url)
		if err := setCmd.Run(); err != nil {
			return fmt.Errorf("failed to update remote: %w", err)
		}
		return nil
	}

	// No remote, add it
	addCmd := exec.Command("git", "-C", dir, "remote", "add", "origin", url)
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}
	return nil
}

// HasRemote checks if the aidb repository has a remote configured
func HasRemote(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	return cmd.Run() == nil
}

// GetRemoteURL returns the origin remote URL if configured
func GetRemoteURL(dir string) string {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// HasUpstream checks if the current branch has an upstream configured
func HasUpstream(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "@{upstream}")
	cmd.Stderr = nil
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "main"
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "main"
	}
	return branch
}

// IsInitialized checks if aidb is initialized
func IsInitialized(cfg *config.Config) bool {
	_, err := os.Stat(cfg.DBDir)
	return err == nil
}
