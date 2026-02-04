package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
		matches, err := filepath.Glob(filepath.Join(cwd, arg))
		if err != nil {
			return fmt.Errorf("invalid glob pattern: %s", arg)
		}
		if len(matches) == 0 {
			// Not a glob, treat as literal path
			files = append(files, filepath.Join(cwd, arg))
		} else {
			files = append(files, matches...)
		}
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

	// Process each file
	for _, srcPath := range files {
		if err := addFile(cfg, srcPath, storageDir, cwd); err != nil {
			printError(fmt.Sprintf("%s: %v", filepath.Base(srcPath), err))
			continue
		}
	}

	return nil
}

func addFile(cfg *config.Config, srcPath, storageDir, cwd string) error {
	info, err := os.Lstat(srcPath)
	if err != nil {
		return fmt.Errorf("file not found")
	}

	// Skip if already a symlink pointing to aidb
	if info.Mode()&os.ModeSymlink != 0 {
		target, _ := os.Readlink(srcPath)
		if filepath.HasPrefix(target, cfg.DBDir) {
			return fmt.Errorf("already tracked")
		}
		return fmt.Errorf("is a symlink")
	}

	// Get relative path from cwd for directory structure
	relPath, err := filepath.Rel(cwd, srcPath)
	if err != nil {
		relPath = filepath.Base(srcPath)
	}

	// Destination in storage
	dstPath := filepath.Join(storageDir, relPath)

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
		printWarning(fmt.Sprintf("%s: git add failed", relPath))
	}

	printSuccess(fmt.Sprintf("Added %s", relPath))
	return nil
}

func addDirectory(cfg *config.Config, srcDir, dstDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(srcDir, path)
		dstPath := filepath.Join(dstDir, relPath)

		// Create parent dirs
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		// Move file
		if err := os.Rename(path, dstPath); err != nil {
			return err
		}

		// Create symlink back
		if err := os.Symlink(dstPath, path); err != nil {
			os.Rename(dstPath, path)
			return err
		}

		// Stage in git
		gitCmd := exec.Command("git", "-C", cfg.DBDir, "add", dstPath)
		gitCmd.Run()

		printSuccess(fmt.Sprintf("Added %s", relPath))
		return nil
	})
}
