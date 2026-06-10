package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

			hash, err := metadata.HashFile(path)
			if err != nil {
				printError(fmt.Sprintf("%s: %v", relPath, err))
				failed++
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

	if failed > 0 {
		return fmt.Errorf("%d of %d failed", failed, total)
	}
	return nil
}

// resolveStoreArg maps a cwd-relative tracked symlink to its store-relative path.
// Args that already resolve under the store pass through unchanged.
func resolveStoreArg(dbDir, pattern string) string {
	if _, err := os.Stat(filepath.Join(dbDir, pattern)); err == nil {
		return pattern
	}
	if target, err := os.Readlink(pattern); err == nil {
		if rel, err := filepath.Rel(dbDir, target); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return pattern
}
