package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var syncMessage string

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Commit and push changes (like git commit + push)",
	Long:  `Commit all changes in the database and push to remote.`,
	RunE:  runSync,
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit changes locally",
	RunE:  runCommit,
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push commits to remote",
	RunE:  runPush,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(pushCmd)

	syncCmd.Flags().StringVarP(&syncMessage, "message", "m", "", "Commit message")
	commitCmd.Flags().StringVarP(&syncMessage, "message", "m", "", "Commit message")
}

func runSync(cmd *cobra.Command, args []string) error {
	if err := runCommit(cmd, args); err != nil {
		return err
	}
	return runPush(cmd, args)
}

func runCommit(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	// Check if there are changes
	gitCmd := exec.Command("git", "-C", cfg.DBDir, "status", "--porcelain")
	out, err := gitCmd.Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	if len(strings.TrimSpace(string(out))) == 0 {
		printInfo("Nothing to commit")
		return nil
	}

	// Stage all changes
	gitCmd = exec.Command("git", "-C", cfg.DBDir, "add", "-A")
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Generate commit message
	msg := syncMessage
	if msg == "" {
		msg = fmt.Sprintf("Update knowledge base %s", time.Now().Format("2006-01-02 15:04"))
	}

	// Commit
	gitCmd = exec.Command("git", "-C", cfg.DBDir, "commit", "-m", msg)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	printSuccess("Committed")
	return nil
}

func runPush(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
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
