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
	out, err := gitCmd.Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		printInfo("Nothing to commit, working tree clean")
		return nil
	}

	fmt.Println("Changes in aidb:")
	fmt.Println()

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		status := line[:2]
		file := strings.TrimSpace(line[2:])

		switch {
		case status[0] == 'A':
			fmt.Printf("  %s new file:   %s\n", colorGreen("+"), file)
		case status[0] == 'M' || status[1] == 'M':
			fmt.Printf("  %s modified:   %s\n", colorYellow("~"), file)
		case status[0] == 'D' || status[1] == 'D':
			fmt.Printf("  %s deleted:    %s\n", colorRed("-"), file)
		case status == "??":
			fmt.Printf("  %s untracked:  %s\n", colorGray("?"), file)
		default:
			fmt.Printf("  %s %s\n", status, file)
		}
	}

	fmt.Println()
	return nil
}

func colorGreen(s string) string {
	return fmt.Sprintf("\033[0;32m%s\033[0m", s)
}

func colorYellow(s string) string {
	return fmt.Sprintf("\033[1;33m%s\033[0m", s)
}

func colorRed(s string) string {
	return fmt.Sprintf("\033[0;31m%s\033[0m", s)
}

func colorGray(s string) string {
	return fmt.Sprintf("\033[0;90m%s\033[0m", s)
}
