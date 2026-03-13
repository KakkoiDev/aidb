package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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
		return fmt.Errorf("aidb not initialized. Run: aidb init")
	}

	if !HasRemote(cfg.DBDir) {
		return fmt.Errorf("no remote configured. Run: aidb init --remote <url>")
	}

	// Ensure pull.rebase is set so even raw `git pull` from ~/.aidb works
	ensureRebaseConfig(cfg.DBDir)

	gitArgs := []string{"-C", cfg.DBDir}

	// Pull with rebase and autostash (handles dirty worktree automatically)
	pullExec := exec.Command("git", append(gitArgs, "pull", "--rebase", "--autostash")...)
	pullExec.Stdout = os.Stdout
	pullExec.Stderr = os.Stderr
	pullErr := pullExec.Run()

	if pullErr != nil {
		// Check if we're stuck in a rebase
		if isRebaseInProgress(cfg.DBDir) {
			printWarning("Rebase conflict detected, aborting rebase")
			abort := exec.Command("git", append(gitArgs, "rebase", "--abort")...)
			abort.Stdout = os.Stdout
			abort.Stderr = os.Stderr
			_ = abort.Run()
			return fmt.Errorf("pull failed: rebase conflict. Resolve manually or force with: cd ~/.aidb && git pull --rebase")
		}
		return fmt.Errorf("git pull failed: %w", pullErr)
	}

	printSuccess("Pulled")
	return nil
}

// ensureRebaseConfig sets pull.rebase=true in the repo config if not already set.
// This prevents the "divergent branches" error when pulling outside of aidb.
func ensureRebaseConfig(dir string) {
	check := exec.Command("git", "-C", dir, "config", "--local", "pull.rebase")
	out, err := check.Output()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		set := exec.Command("git", "-C", dir, "config", "--local", "pull.rebase", "true")
		_ = set.Run()
	}
}

// isRebaseInProgress checks if a rebase is currently in progress
func isRebaseInProgress(dir string) bool {
	for _, subdir := range []string{"rebase-merge", "rebase-apply"} {
		gitDir := dir + "/.git/" + subdir
		if _, err := os.Stat(gitDir); err == nil {
			return true
		}
	}
	return false
}
