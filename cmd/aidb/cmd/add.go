package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <filename>",
	Short: "Create new file in database with symlink",
	Long: `Create a new file in the database and symlink it to current directory.

Examples:
  aidb add notes.md
  aidb add analysis.qmd`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var trackCmd = &cobra.Command{
	Use:   "track <filename>",
	Short: "Move existing file to database with symlink back",
	Long: `Move an existing file to the database and create a symlink in its place.

Examples:
  aidb track existing-doc.md`,
	Args: cobra.ExactArgs(1),
	RunE: runTrack,
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(trackCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	filename := args[0]

	cfg, err := config.New()
	if err != nil {
		return err
	}

	project, task, err := cfg.GetProjectFromCwd()
	if err != nil {
		return err
	}
	cfg.SetProject(project, task)

	if err := cfg.EnsureProjectDir(); err != nil {
		return err
	}

	// Create file in database
	dbPath := cfg.GetFilePath(filename)
	if fileExists(dbPath) {
		return fmt.Errorf("file already exists in database: %s", dbPath)
	}

	// Create empty file
	if err := os.WriteFile(dbPath, []byte(""), 0644); err != nil {
		return err
	}

	// Create symlink in current directory
	cwd, _ := os.Getwd()
	linkPath := filepath.Join(cwd, filename)
	if err := createSymlink(dbPath, linkPath); err != nil {
		return err
	}

	printSuccess(fmt.Sprintf("Created %s -> %s", filename, dbPath))
	return nil
}

func runTrack(cmd *cobra.Command, args []string) error {
	filename := args[0]

	cfg, err := config.New()
	if err != nil {
		return err
	}

	project, task, err := cfg.GetProjectFromCwd()
	if err != nil {
		return err
	}
	cfg.SetProject(project, task)

	if err := cfg.EnsureProjectDir(); err != nil {
		return err
	}

	cwd, _ := os.Getwd()
	srcPath := filepath.Join(cwd, filename)

	// Check source exists and is not a symlink
	info, err := os.Lstat(srcPath)
	if err != nil {
		return fmt.Errorf("file not found: %s", srcPath)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("file is already a symlink: %s", srcPath)
	}

	// Move to database
	dbPath := cfg.GetFilePath(filename)
	if err := os.Rename(srcPath, dbPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	// Create symlink back
	if err := createSymlink(dbPath, srcPath); err != nil {
		return err
	}

	printSuccess(fmt.Sprintf("Tracked %s -> %s", filename, dbPath))
	return nil
}
