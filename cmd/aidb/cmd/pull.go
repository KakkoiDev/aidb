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

	gitArgs := []string{"-C", cfg.DBDir}

	// Stash unstaged changes before pulling
	stash := exec.Command("git", append(gitArgs, "stash", "--include-untracked")...)
	stash.Stdout = os.Stdout
	stash.Stderr = os.Stderr
	if err := stash.Run(); err != nil {
		return fmt.Errorf("git stash failed: %w", err)
	}

	pullCmd := exec.Command("git", append(gitArgs, "pull", "--rebase")...)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	pullErr := pullCmd.Run()

	// Pop stash regardless of pull result
	pop := exec.Command("git", append(gitArgs, "stash", "pop")...)
	pop.Stdout = os.Stdout
	pop.Stderr = os.Stderr
	_ = pop.Run() // ignore error if stash was empty ("No stash entries")

	if pullErr != nil {
		return fmt.Errorf("git pull failed: %w", pullErr)
	}

	printSuccess("Pulled")
	return nil
}
