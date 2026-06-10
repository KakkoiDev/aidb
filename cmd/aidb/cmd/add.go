package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <file|dir|glob>",
	Short: "Move file to ~/.aidb, create symlink, stage in git",
	Long: `Move file(s) to the aidb database, create symlinks back, and stage in git.

Examples:
  aidb add TASK.md
  aidb add *.md
  aidb add docs/`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Expand globs
	var files []string
	for _, arg := range args {
		pattern := resolveArg(cwd, arg)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid glob pattern: %s", arg)
		}
		if len(matches) == 0 {
			// Not a glob, treat as literal path
			files = append(files, pattern)
		} else {
			files = append(files, matches...)
		}
	}

	// In a git repo the store path is anchored at the repo toplevel, not the cwd
	top := config.GetGitToplevel(cwd)
	anchor := cwd
	if top != "" {
		anchor = top
	}

	// Ensure base DB dir exists with git
	if err := cfg.EnsureDBDir(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Ensure storage dir exists
	storageDir, err := cfg.EnsureStorageDir()
	if err != nil {
		return fmt.Errorf("failed to create storage dir: %w", err)
	}

	// Git projects are keyed by basename; refuse a colliding repo's add
	if top != "" {
		if err := pinOrigin(cfg.DBDir, filepath.Dir(storageDir), top); err != nil {
			return err
		}
	}

	// Process each file
	failed := 0
	for _, srcPath := range files {
		if err := addFile(cfg, srcPath, storageDir, anchor); err != nil {
			printError(fmt.Sprintf("%s: %v", filepath.Base(srcPath), err))
			failed++
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d failed", failed, len(files))
	}
	return nil
}

func addFile(cfg *config.Config, srcPath, storageDir, anchor string) error {
	info, err := os.Lstat(srcPath)
	if err != nil {
		return fmt.Errorf("file not found")
	}

	// Already a symlink pointing into the store: re-stage its current content
	if info.Mode()&os.ModeSymlink != 0 {
		target, _ := os.Readlink(srcPath)
		if withinDir(cfg.DBDir, target) {
			gitCmd := exec.Command("git", "-C", cfg.DBDir, "add", target)
			if err := gitCmd.Run(); err != nil {
				return fmt.Errorf("failed to re-stage: %w", err)
			}
			printSuccess(fmt.Sprintf("Re-staged %s", filepath.Base(target)))
			return nil
		}
		return fmt.Errorf("is a symlink")
	}

	// Normalize symlinked parents (e.g. /tmp -> /private/tmp) so Rel against
	// the git toplevel compares physical paths
	if resolved, err := filepath.EvalSymlinks(srcPath); err == nil {
		srcPath = resolved
	}

	// Get relative path from the anchor for directory structure
	relPath, err := filepath.Rel(anchor, srcPath)
	if err != nil {
		relPath = filepath.Base(srcPath)
	}

	// Destination in storage
	dstPath := filepath.Join(storageDir, relPath)
	if !withinDir(storageDir, dstPath) {
		return fmt.Errorf("outside project namespace")
	}

	// Create parent dirs in storage
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}

	// Handle directory
	if info.IsDir() {
		return addDirectory(cfg, srcPath, dstPath)
	}

	// Check if destination already exists
	if _, err := os.Stat(dstPath); err == nil {
		return fmt.Errorf("already exists in database")
	}

	// Move file to storage
	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to move: %w", err)
	}

	// Create symlink back
	if err := os.Symlink(dstPath, srcPath); err != nil {
		// Rollback: move file back
		os.Rename(dstPath, srcPath)
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Stage in git
	gitCmd := exec.Command("git", "-C", cfg.DBDir, "add", dstPath)
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("moved and symlinked, but git add failed: %w", err)
	}

	printSuccess(fmt.Sprintf("Added %s", relPath))
	return nil
}

func addDirectory(cfg *config.Config, srcDir, dstDir string) error {
	var errs []string
	walkErr := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(srcDir, path)
		dstPath := filepath.Join(dstDir, relPath)

		// Create parent dirs
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", relPath, err))
			return nil
		}

		// Move file
		if err := os.Rename(path, dstPath); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", relPath, err))
			return nil
		}

		// Create symlink back
		if err := os.Symlink(dstPath, path); err != nil {
			os.Rename(dstPath, path)
			errs = append(errs, fmt.Sprintf("%s: %v", relPath, err))
			return nil
		}

		// Stage in git
		gitCmd := exec.Command("git", "-C", cfg.DBDir, "add", dstPath)
		if err := gitCmd.Run(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: git add failed", relPath))
			return nil
		}

		printSuccess(fmt.Sprintf("Added %s", relPath))
		return nil
	})
	if walkErr != nil {
		return walkErr
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}
