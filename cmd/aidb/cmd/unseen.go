package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/KakkoiDev/aidb/internal/metadata"
	"github.com/spf13/cobra"
)

var unseenCmd = &cobra.Command{
	Use:   "unseen <file|glob>",
	Short: "Mark file(s) for re-processing",
	Long: `Mark one or more files as unseen for re-processing by AI agents.

Examples:
  aidb unseen TASK.md
  aidb unseen "project/main/*.md"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runUnseen,
}

func init() {
	rootCmd.AddCommand(unseenCmd)
}

func runUnseen(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	meta, err := metadata.New(cfg.DBDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	count := 0
	total := 0
	failed := 0
	for _, pattern := range args {
		pattern = resolveStoreArg(cfg.DBDir, pattern)
		matches, err := filepath.Glob(filepath.Join(cfg.DBDir, pattern))
		if err != nil {
			printError(fmt.Sprintf("invalid pattern: %s", pattern))
			total++
			failed++
			continue
		}

		if len(matches) == 0 {
			// Try as literal path
			matches = []string{filepath.Join(cfg.DBDir, pattern)}
		}

		for _, path := range matches {
			total++
			relPath, err := filepath.Rel(cfg.DBDir, path)
			if err != nil {
				printError(fmt.Sprintf("%s: %v", path, err))
				failed++
				continue
			}

			// Unknown to metadata AND absent from the store = nothing to mark
			if meta.GetInfo(relPath) == nil {
				if _, err := os.Stat(path); err != nil {
					printError(fmt.Sprintf("%s: not tracked", relPath))
					failed++
					continue
				}
			}

			meta.MarkUnseen(relPath)
			printSuccess(fmt.Sprintf("Marked unseen: %s", relPath))
			count++
		}
	}

	if count > 0 {
		if err := meta.Save(); err != nil {
			return fmt.Errorf("failed to save metadata: %w", err)
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d failed", failed, total)
	}
	return nil
}
