package cmd

import (
	"fmt"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "Show or set configuration",
	Long: `Show or set aidb configuration values.

With no arguments, shows all configuration.
With one argument, shows that key's value.
With two arguments, sets the key to the value.

Config file: ~/.config/aidb/config.yaml

Keys:
  db.path     Database location (default ~/.aidb). Changing it does NOT
              migrate existing files or symlinks.
  git.remote  Origin remote of the database repository.

backup.enabled is read-only here: it reports the launchd agent state,
managed by 'aidb backup enable/disable'.

Examples:
  aidb config              # Show all config
  aidb config db.path      # Show db.path value
  aidb config db.path /custom/path  # Set db.path`,
	Args: cobra.MaximumNArgs(2),
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	userCfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// No args: show all config
	if len(args) == 0 {
		fmt.Println("# Current configuration")
		fmt.Println()
		fmt.Printf("db.path = %s\n", cfg.DBDir)
		fmt.Printf("backup.enabled = %v\n", backupPlistInstalled(cfg))
		fmt.Printf("git.remote = %s\n", GetRemoteURL(cfg.DBDir))
		fmt.Println()
		fmt.Printf("# Config file: %s\n", config.UserConfigPath())
		return nil
	}

	key := args[0]

	// One arg: show specific key
	if len(args) == 1 {
		switch key {
		case "db.path":
			fmt.Println(cfg.DBDir)
		case "backup.enabled":
			fmt.Println(backupPlistInstalled(cfg))
		case "git.remote":
			fmt.Println(GetRemoteURL(cfg.DBDir))
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
		return nil
	}

	// Two args: set key
	value := args[1]
	switch key {
	case "db.path":
		userCfg.DB.Path = value
	case "backup.enabled":
		return fmt.Errorf("backup.enabled is managed by 'aidb backup enable/disable'")
	case "git.remote":
		if err := configureRemote(cfg.DBDir, value); err != nil {
			return fmt.Errorf("failed to configure remote: %w", err)
		}
		printSuccess(fmt.Sprintf("Set %s = %s", key, value))
		return nil
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	if err := config.SaveUserConfig(userCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	printSuccess(fmt.Sprintf("Set %s = %s", key, value))
	return nil
}
