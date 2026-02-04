package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagJSON    bool
	flagQuiet   bool
	flagNoColor bool
	flagDebug   bool
)

var rootCmd = &cobra.Command{
	Use:   "aidb",
	Short: "AI Knowledge Database - centralized knowledge file management",
	Long: `aidb centralizes knowledge files (TASK.md, MEMO.md, LEARN.md) for AI agent workflows.

All files are stored in ~/.aidb with symlinks back to original locations.
Built-in git versioning provides disaster recovery.

Core Commands:
  add       Move file to ~/.aidb, create symlink, stage in git
  commit    Commit staged changes
  status    Show staged/unstaged changes
  push      Push commits to remote
  pull      Pull changes from remote
  remove    Untrack file, restore to original location

AI Agent Commands:
  list      List tracked files with metadata
  seen      Mark file(s) as processed by AI
  unseen    Mark file(s) for re-processing

Configuration:
  config    Show or set configuration
  backup    Enable/disable automatic backup (hourly commit + push)

Examples:
  aidb add TASK.md         # Start tracking a file
  aidb commit "Add notes"  # Commit changes
  aidb list --unseen       # Show files needing AI processing`,
	Version: "0.2.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Global flags (clig.dev compliant)
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&flagDebug, "debug", "d", false, "Show debug output")
}

// Helper functions for colored output
func printInfo(msg string) {
	if flagQuiet {
		return
	}
	if flagNoColor {
		fmt.Fprintf(os.Stdout, "[INFO] %s\n", msg)
	} else {
		fmt.Fprintf(os.Stdout, "\033[0;34m[INFO]\033[0m %s\n", msg)
	}
}

func printSuccess(msg string) {
	if flagQuiet {
		return
	}
	if flagNoColor {
		fmt.Fprintf(os.Stdout, "✓ %s\n", msg)
	} else {
		fmt.Fprintf(os.Stdout, "\033[0;32m✓\033[0m %s\n", msg)
	}
}

func printError(msg string) {
	if flagNoColor {
		fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "\033[0;31m✗\033[0m %s\n", msg)
	}
}

func printWarning(msg string) {
	if flagQuiet {
		return
	}
	if flagNoColor {
		fmt.Fprintf(os.Stdout, "! %s\n", msg)
	} else {
		fmt.Fprintf(os.Stdout, "\033[1;33m!\033[0m %s\n", msg)
	}
}

func printDebug(msg string) {
	if !flagDebug {
		return
	}
	if flagNoColor {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "\033[0;36m[DEBUG]\033[0m %s\n", msg)
	}
}
