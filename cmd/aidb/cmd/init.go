package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var autoInit bool

var initCmd = &cobra.Command{
	Use:   "init [project/task]",
	Short: "Initialize aidb for current project",
	Long: `Initialize TASK.md and MEMO.md symlinks for the current directory.

If no argument is provided, auto-detects project name from git repo
and task name from git branch.

Examples:
  aidb init                  # Auto-detect from git
  aidb init myproject/main   # Explicit project/task
  aidb init --auto           # Auto-initialize if files missing`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&autoInit, "auto", false, "Auto-initialize if TASK.md or MEMO.md are missing")
}

func runInit(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	var project, task string
	if len(args) > 0 {
		// Parse project/task from argument
		parts := filepath.SplitList(args[0])
		if len(parts) == 0 {
			parts = []string{args[0]}
		}
		// Handle "project/task" format
		if idx := filepath.Base(args[0]); idx != args[0] {
			project = filepath.Dir(args[0])
			task = idx
		} else {
			project = args[0]
			task = "main"
		}
	} else {
		// Auto-detect from git
		project, task, err = cfg.GetProjectFromCwd()
		if err != nil {
			return fmt.Errorf("failed to detect project: %w", err)
		}
	}

	cfg.SetProject(project, task)

	// Check if already initialized (for --auto mode)
	cwd, _ := os.Getwd()
	taskExists := fileExists(filepath.Join(cwd, "TASK.md"))
	memoExists := fileExists(filepath.Join(cwd, "MEMO.md"))

	if autoInit && taskExists && memoExists {
		printInfo("Already initialized")
		return nil
	}

	// Create project directory
	if err := cfg.EnsureProjectDir(); err != nil {
		return fmt.Errorf("failed to create project dir: %w", err)
	}

	// Create files if they don't exist in database
	if err := ensureFile(cfg.GetFilePath("TASK.md"), defaultTaskContent()); err != nil {
		return err
	}
	if err := ensureFile(cfg.GetFilePath("MEMO.md"), defaultMemoContent()); err != nil {
		return err
	}

	// Create symlinks in current directory
	if err := createSymlink(cfg.GetFilePath("TASK.md"), filepath.Join(cwd, "TASK.md")); err != nil {
		return err
	}
	if err := createSymlink(cfg.GetFilePath("MEMO.md"), filepath.Join(cwd, "MEMO.md")); err != nil {
		return err
	}

	printSuccess(fmt.Sprintf("Initialized %s/%s", project, task))
	printInfo(fmt.Sprintf("Database: %s", cfg.ProjectDir))
	return nil
}

func fileExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func ensureFile(path, content string) error {
	if fileExists(path) {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func createSymlink(src, dst string) error {
	// Remove existing file/symlink
	os.Remove(dst)
	return os.Symlink(src, dst)
}

func defaultTaskContent() string {
	return `# Task Plan

## Status: Planning

## Tasks
- [ ] Define requirements
- [ ] Implementation
- [ ] Testing

## Notes
`
}

func defaultMemoContent() string {
	return `# Project Memo

## Overview

## Key Findings

## Architecture

## References
`
}
