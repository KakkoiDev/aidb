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
	Short: "Centralized file management with git versioning",
	Long: `aidb stores files in ~/.aidb with symlinks back to original locations.

Commands:
  aidb init [--remote <url>]   Initialize database
  aidb add <file>              Track file
  aidb remove <file>           Untrack file
  aidb list [--unseen]         List tracked files
  aidb seen/unseen <file>      Mark file status
  aidb status                  Show changes
  aidb commit <msg>            Commit changes
  aidb push/pull               Sync with remote

Usage modes (all opt-in):
  CLI only         Just this tool for manual knowledge management
  + Skill          Add SKILL.md to ~/.claude/skills/aidb/ for AI prompting
  + Agent          Add agents/aidb.md to ~/.claude/agents/ for automation`,
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
