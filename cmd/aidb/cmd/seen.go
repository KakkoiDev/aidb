package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/KakkoiDev/aidb/internal/metadata"
	"github.com/spf13/cobra"
)

var seenCmd = &cobra.Command{
	Use:   "seen <file|glob>",
	Short: "Mark file(s) as processed by AI",
	Long: `Mark one or more files as seen/processed by AI agents.

The current file hash is stored so changes can be detected later.

Examples:
  aidb seen TASK.md
  aidb seen "project/main/*.md"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSeen,
}

func init() {
	rootCmd.AddCommand(seenCmd)
}

func runSeen(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	meta, err := metadata.New(cfg.DBDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	count := 0
	for _, pattern := range args {
		matches, err := filepath.Glob(filepath.Join(cfg.DBDir, pattern))
		if err != nil {
			printError(fmt.Sprintf("invalid pattern: %s", pattern))
			continue
		}

		if len(matches) == 0 {
			// Try as literal path
			matches = []string{filepath.Join(cfg.DBDir, pattern)}
		}

		for _, path := range matches {
			relPath, err := filepath.Rel(cfg.DBDir, path)
			if err != nil {
				continue
			}

			hash, err := metadata.HashFile(path)
			if err != nil {
				printError(fmt.Sprintf("%s: %v", relPath, err))
				continue
			}

			meta.MarkSeen(relPath, hash)
			printSuccess(fmt.Sprintf("Marked seen: %s", relPath))
			count++
		}
	}

	if count > 0 {
		if err := meta.Save(); err != nil {
			return fmt.Errorf("failed to save metadata: %w", err)
		}
	}

	return nil
}
