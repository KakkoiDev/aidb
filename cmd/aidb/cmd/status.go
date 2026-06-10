package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show staged/unstaged changes",
	Long:  `Show the status of tracked files in the aidb database.`,
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

	// Check if database directory exists
	if _, err := os.Stat(cfg.DBDir); os.IsNotExist(err) {
		printInfo("aidb not initialized (run 'aidb add' first)")
		return nil
	}

	// Run git status
	gitCmd := exec.Command("git", "-C", cfg.DBDir, "status", "--short")
	statusOut, err := gitCmd.Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	trimmed := strings.TrimSpace(string(statusOut))
	if trimmed == "" {
		printInfo("Nothing to commit, working tree clean")
		return nil
	}

	fmt.Println("Changes in aidb:")
	fmt.Println()

	lines := strings.Split(trimmed, "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		file := strings.TrimSpace(line[2:])

		switch {
		case status[0] == 'A':
			fmt.Printf("  %s new file:   %s\n", out.Colorize("0;32", "+"), file)
		case status[0] == 'M' || status[1] == 'M':
			fmt.Printf("  %s modified:   %s\n", out.Colorize("1;33", "~"), file)
		case status[0] == 'D' || status[1] == 'D':
			fmt.Printf("  %s deleted:    %s\n", out.Colorize("0;31", "-"), file)
		case status == "??":
			fmt.Printf("  %s untracked:  %s\n", out.Colorize("0;90", "?"), file)
		default:
			fmt.Printf("  %s %s\n", status, file)
		}
	}

	fmt.Println()
	return nil
}
