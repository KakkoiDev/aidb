package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <file>",
	Short: "Untrack file, restore to original location",
	Long: `Remove a file from aidb tracking and restore it to its original location.

The file is moved back from the database and the symlink is replaced with the actual file.
The file remains in git history for recovery.

Examples:
  aidb remove TASK.md`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	filename := args[0]

	cfg, err := config.New()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	linkPath := filepath.Join(cwd, filename)

	// Check if it's a symlink
	info, err := os.Lstat(linkPath)
	if err != nil {
		return fmt.Errorf("file not found: %s", filename)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("not a tracked file (not a symlink): %s", filename)
	}

	// Get symlink target
	target, err := os.Readlink(linkPath)
	if err != nil {
		return fmt.Errorf("failed to read symlink: %w", err)
	}

	// Verify target is in aidb
	if !strings.HasPrefix(target, cfg.DBDir) {
		return fmt.Errorf("file is not tracked by aidb: %s", filename)
	}

	// Check if source file exists
	if _, err := os.Stat(target); err != nil {
		return fmt.Errorf("database file missing: %s", target)
	}

	// Remove symlink
	if err := os.Remove(linkPath); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	// Move file back
	if err := os.Rename(target, linkPath); err != nil {
		// Try to restore symlink on failure
		os.Symlink(target, linkPath)
		return fmt.Errorf("failed to restore file: %w", err)
	}

	// Remove from git (--cached keeps history)
	gitCmd := exec.Command("git", "-C", cfg.DBDir, "rm", "--cached", target)
	gitCmd.Run() // Ignore error, file might not be staged

	printSuccess(fmt.Sprintf("Removed %s from tracking", filename))
	return nil
}
