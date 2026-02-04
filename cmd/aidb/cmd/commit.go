package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit <message>",
	Short: "Commit staged changes",
	Long: `Commit staged changes to the aidb repository.

Examples:
  aidb commit "Add project notes"
  aidb commit "Update TASK.md with new requirements"`,
	Args: cobra.ExactArgs(1),
	RunE: runCommit,
}

func init() {
	rootCmd.AddCommand(commitCmd)
}

func runCommit(cmd *cobra.Command, args []string) error {
	message := args[0]
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}

	// Check if database directory exists
	if _, err := os.Stat(cfg.DBDir); os.IsNotExist(err) {
		return fmt.Errorf("aidb not initialized (run 'aidb add' first)")
	}

	// Check for staged changes
	gitCmd := exec.Command("git", "-C", cfg.DBDir, "diff", "--cached", "--quiet")
	if err := gitCmd.Run(); err == nil {
		printInfo("Nothing staged to commit")
		return nil
	}

	// Commit
	gitCmd = exec.Command("git", "-C", cfg.DBDir, "commit", "-m", message)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	printSuccess("Committed")
	return nil
}
