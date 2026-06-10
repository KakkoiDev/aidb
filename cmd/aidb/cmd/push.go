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

	if err := pushWithUpstream(cfg.DBDir); err != nil {
		return err
	}

	printSuccess("Pushed")
	return nil
}

// pushWithUpstream pushes, setting the upstream on first push.
// With an upstream it pulls (rebase) first and retries once on rejection.
func pushWithUpstream(dbDir string) error {
	if !HasUpstream(dbDir) {
		branch := GetCurrentBranch(dbDir)
		return gitPush(dbDir, "-u", "origin", branch)
	}

	if err := pullRebase(dbDir); err != nil {
		return err
	}
	if err := gitPush(dbDir); err == nil {
		return nil
	}
	// Rejected: a push raced ours between the pull and the push
	if err := pullRebase(dbDir); err != nil {
		return err
	}
	return gitPush(dbDir)
}

func gitPush(dbDir string, extraArgs ...string) error {
	args := append([]string{"-C", dbDir, "push"}, extraArgs...)
	gitCmd := exec.Command("git", args...)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}
	return nil
}
