package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aidb",
	Short: "AI Knowledge Database - manage your AI development knowledge",
	Long: `aidb is a git-like tool for managing AI development knowledge.
It tracks TASK.md, MEMO.md, and LEARN.md files with versioning support.

Examples:
  aidb init          Initialize project in ~/.aidb
  aidb init --auto   Auto-detect project and initialize
  aidb status        Show learning status (like git status)
  aidb sync          Commit and push changes`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

// Helper functions for colored output
func printInfo(msg string) {
	fmt.Fprintf(os.Stdout, "\033[0;34m[INFO]\033[0m %s\n", msg)
}

func printSuccess(msg string) {
	fmt.Fprintf(os.Stdout, "\033[0;32m✓\033[0m %s\n", msg)
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "\033[0;31m✗\033[0m %s\n", msg)
}

func printWarning(msg string) {
	fmt.Fprintf(os.Stdout, "\033[1;33m!\033[0m %s\n", msg)
}
