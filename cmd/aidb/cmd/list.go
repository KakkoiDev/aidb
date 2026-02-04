package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/KakkoiDev/aidb/internal/metadata"
	"github.com/spf13/cobra"
)

var (
	listUnseen bool
	listJSON   bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracked files with metadata",
	Long: `List all tracked files in the aidb database with their metadata.

Examples:
  aidb list           # List all files
  aidb list --unseen  # List only unseen files
  aidb list --json    # Output as JSON`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVar(&listUnseen, "unseen", false, "Show only unseen files")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
}

type FileEntry struct {
	Path     string `json:"path"`
	Seen     bool   `json:"seen"`
	Hash     string `json:"hash,omitempty"`
	SeenAt   string `json:"seenAt,omitempty"`
	Modified bool   `json:"modified,omitempty"`
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	if _, err := os.Stat(cfg.DBDir); os.IsNotExist(err) {
		if listJSON {
			fmt.Println("[]")
		} else {
			printInfo("No tracked files")
		}
		return nil
	}

	meta, err := metadata.New(cfg.DBDir)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	var entries []FileEntry

	err = filepath.Walk(cfg.DBDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// Skip .git directory only (allow .aidb root and other dirs)
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip metadata file
		if info.Name() == ".metadata.json" {
			return nil
		}

		relPath, _ := filepath.Rel(cfg.DBDir, path)

		// Get current hash
		currentHash, _ := metadata.HashFile(path)

		// Check seen status
		seen := meta.IsSeen(relPath, currentHash)
		fileInfo := meta.GetInfo(relPath)

		entry := FileEntry{
			Path: relPath,
			Seen: seen,
		}

		if fileInfo != nil {
			entry.Hash = fileInfo.Hash
			if !fileInfo.SeenAt.IsZero() {
				entry.SeenAt = fileInfo.SeenAt.Format("2006-01-02T15:04:05Z")
			}
			// Check if modified since last seen
			if fileInfo.Hash != "" && fileInfo.Hash != currentHash {
				entry.Modified = true
			}
		}

		// Filter unseen if requested
		if listUnseen && seen {
			return nil
		}

		entries = append(entries, entry)
		return nil
	})

	if err != nil {
		return err
	}

	if listJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}

	if len(entries) == 0 {
		if listUnseen {
			printInfo("No unseen files")
		} else {
			printInfo("No tracked files")
		}
		return nil
	}

	for _, e := range entries {
		status := colorGray("○")
		if e.Seen {
			status = colorGreen("●")
		}
		if e.Modified {
			status = colorYellow("◐")
		}

		fmt.Printf("  %s %s\n", status, e.Path)
	}

	return nil
}
