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
		return fmt.Errorf("aidb not initialized")
	}

	gitCmd := exec.Command("git", "-C", cfg.DBDir, "push")
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	printSuccess("Pushed")
	return nil
}
