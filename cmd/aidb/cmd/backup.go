package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup <enable|disable|status>",
	Short: "Manage automatic backup",
	Long: `Enable or disable automatic hourly backup (commit + push).

Examples:
  aidb backup enable   # Enable hourly backup
  aidb backup disable  # Disable backup
  aidb backup status   # Show backup configuration`,
	Args: cobra.ExactArgs(1),
	RunE: runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func runBackup(cmd *cobra.Command, args []string) error {
	action := args[0]

	switch action {
	case "enable":
		return enableBackup()
	case "disable":
		return disableBackup()
	case "status":
		return backupStatus()
	default:
		return fmt.Errorf("unknown action: %s (use enable, disable, or status)", action)
	}
}

func enableBackup() error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("automatic backup only supported on macOS (launchd)")
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}

	// Find aidb binary
	aidbPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find aidb binary: %w", err)
	}

	// Create LaunchAgent plist
	plistPath := filepath.Join(cfg.HomeDir, "Library", "LaunchAgents", "com.aidb.backup.plist")

	plistTemplate := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.aidb.backup</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.AidbPath}}</string>
        <string>backup-run</string>
    </array>
    <key>StartInterval</key>
    <integer>3600</integer>
    <key>RunAtLoad</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>
</dict>
</plist>
`

	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return err
	}

	// Ensure LaunchAgents directory exists
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(plistPath)
	if err != nil {
		return err
	}
	defer f.Close()

	logPath := filepath.Join(cfg.HomeDir, ".aidb", "backup.log")
	if err := tmpl.Execute(f, map[string]string{
		"AidbPath": aidbPath,
		"LogPath":  logPath,
	}); err != nil {
		return err
	}

	// Load the launch agent
	exec.Command("launchctl", "unload", plistPath).Run() // Ignore error if not loaded
	if err := exec.Command("launchctl", "load", plistPath).Run(); err != nil {
		return fmt.Errorf("failed to load launch agent: %w", err)
	}

	printSuccess("Backup enabled (hourly commit + push)")
	printInfo(fmt.Sprintf("Log file: %s", logPath))
	return nil
}

func disableBackup() error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("automatic backup only supported on macOS (launchd)")
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}

	plistPath := filepath.Join(cfg.HomeDir, "Library", "LaunchAgents", "com.aidb.backup.plist")

	// Unload and remove
	exec.Command("launchctl", "unload", plistPath).Run()
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist: %w", err)
	}

	printSuccess("Backup disabled")
	return nil
}

func backupStatus() error {
	if runtime.GOOS != "darwin" {
		printInfo("Automatic backup only supported on macOS")
		return nil
	}

	cfg, err := config.New()
	if err != nil {
		return err
	}

	plistPath := filepath.Join(cfg.HomeDir, "Library", "LaunchAgents", "com.aidb.backup.plist")

	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		printInfo("Backup is disabled")
		return nil
	}

	// Check if loaded
	out, _ := exec.Command("launchctl", "list", "com.aidb.backup").Output()
	if len(out) > 0 {
		printSuccess("Backup is enabled and running")
	} else {
		printWarning("Backup plist exists but not loaded")
	}

	logPath := filepath.Join(cfg.HomeDir, ".aidb", "backup.log")
	if info, err := os.Stat(logPath); err == nil {
		printInfo(fmt.Sprintf("Last log update: %s", info.ModTime().Format(time.RFC3339)))
	}

	return nil
}

// Internal command for backup execution
var backupRunCmd = &cobra.Command{
	Use:    "backup-run",
	Hidden: true,
	RunE:   runBackupExec,
}

func init() {
	rootCmd.AddCommand(backupRunCmd)
}

func runBackupExec(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	// Check for changes
	gitCmd := exec.Command("git", "-C", cfg.DBDir, "status", "--porcelain")
	out, err := gitCmd.Output()
	if err != nil {
		return err
	}

	if len(out) == 0 {
		fmt.Printf("[%s] No changes to backup\n", time.Now().Format(time.RFC3339))
		return nil
	}

	// Stage all
	if err := exec.Command("git", "-C", cfg.DBDir, "add", "-A").Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Commit
	msg := fmt.Sprintf("Auto-backup %s", time.Now().Format("2006-01-02 15:04"))
	if err := exec.Command("git", "-C", cfg.DBDir, "commit", "-m", msg).Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	// Push
	if err := exec.Command("git", "-C", cfg.DBDir, "push").Run(); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	fmt.Printf("[%s] Backup completed\n", time.Now().Format(time.RFC3339))
	return nil
}
