package cmd

import (
	"runtime/debug"

	"github.com/KakkoiDev/aidb/internal/output"
	"github.com/spf13/cobra"
)

var version = "dev"

var (
	flagQuiet   bool
	flagNoColor bool
)

var out = output.Default()

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
  aidb push/pull               Sync with remote`,
	Version: version,
	// Runtime failures print the error only; usage is for usage errors (clig.dev)
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		out = output.New(output.Options{Quiet: flagQuiet, NoColor: flagNoColor})
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" && info.Main.Version != "" {
			version = info.Main.Version
		}
	}
	rootCmd.Version = version
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Global flags (clig.dev compliant)
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
}

// Helper functions for colored output
func printInfo(msg string)    { out.Info(msg) }
func printSuccess(msg string) { out.Success(msg) }
func printError(msg string)   { out.Error(msg) }
func printWarning(msg string) { out.Warning(msg) }
