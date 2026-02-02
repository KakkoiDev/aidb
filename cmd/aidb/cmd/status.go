package cmd

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show learning status (like git status)",
	Long:  `Scan LEARN.md files and show which need updating based on content changes.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	// Find all LEARN.md files
	var learnFiles []string
	err = filepath.Walk(cfg.DBDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.Name() == "LEARN.md" {
			learnFiles = append(learnFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(learnFiles) == 0 {
		printInfo("No LEARN.md files found")
		return nil
	}

	fmt.Println("Learning Status:")
	fmt.Println()

	needsUpdate := 0
	for _, learnPath := range learnFiles {
		dir := filepath.Dir(learnPath)
		relPath, _ := filepath.Rel(cfg.DBDir, dir)

		// Get content hash of directory (excluding LEARN.md)
		contentHash := hashDirectory(dir, "LEARN.md")

		// Check if LEARN.md contains the hash
		learnContent, err := os.ReadFile(learnPath)
		if err != nil {
			continue
		}

		hashLine := fmt.Sprintf("<!-- hash:%s -->", contentHash)
		hasHash := strings.Contains(string(learnContent), hashLine)

		// Get file mod time
		info, _ := os.Stat(learnPath)
		modTime := info.ModTime().Format("2006-01-02")

		if hasHash {
			fmt.Printf("  \033[0;32m✓\033[0m %s (up to date, %s)\n", relPath, modTime)
		} else {
			fmt.Printf("  \033[1;33m!\033[0m %s (needs update, %s)\n", relPath, modTime)
			needsUpdate++
		}
	}

	fmt.Println()
	if needsUpdate > 0 {
		fmt.Printf("%d file(s) need updating\n", needsUpdate)
	} else {
		printSuccess("All LEARN.md files up to date")
	}

	return nil
}

func hashDirectory(dir, exclude string) string {
	h := md5.New()
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if info.Name() == exclude {
			return nil
		}
		// Add file path and mod time to hash
		relPath, _ := filepath.Rel(dir, path)
		io.WriteString(h, relPath)
		io.WriteString(h, info.ModTime().Format(time.RFC3339))
		return nil
	})
	return fmt.Sprintf("%x", h.Sum(nil))[:8]
}
