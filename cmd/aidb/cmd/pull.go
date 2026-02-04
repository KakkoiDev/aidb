package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull changes from remote",
	Long:  `Pull changes from the remote repository.`,
	RunE:  runPull,
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

func runPull(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	if _, err := os.Stat(cfg.DBDir); os.IsNotExist(err) {
		return fmt.Errorf("aidb not initialized")
	}

	gitCmd := exec.Command("git", "-C", cfg.DBDir, "pull")
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	printSuccess("Pulled")
	return nil
}
