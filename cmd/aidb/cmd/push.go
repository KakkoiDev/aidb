package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push commits to remote",
	Long:  `Push all local commits to the remote repository.`,
	RunE:  runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	if _, err := os.Stat(cfg.DBDir); os.IsNotExist(err) {
		return fmt.Errorf("aidb not initialized. Run: aidb init")
	}

	// Check if remote is configured
	if !HasRemote(cfg.DBDir) {
		return fmt.Errorf("no remote configured. Run: aidb init --remote <url>")
	}

	// Determine push args
	pushArgs := []string{"-C", cfg.DBDir, "push"}

	// Use -u flag if no upstream is set
	if !HasUpstream(cfg.DBDir) {
		branch := GetCurrentBranch(cfg.DBDir)
		pushArgs = []string{"-C", cfg.DBDir, "push", "-u", "origin", branch}
	}

	gitCmd := exec.Command("git", pushArgs...)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	printSuccess("Pushed")
	return nil
}
