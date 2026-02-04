package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KakkoiDev/aidb/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "Show or set configuration",
	Long: `Show or set aidb configuration values.

With no arguments, shows all configuration.
With one argument, shows that key's value.
With two arguments, sets the key to the value.

Config file: ~/.config/aidb/config.yaml

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

type UserConfig struct {
	DB struct {
		Path string `yaml:"path,omitempty"`
	} `yaml:"db,omitempty"`
	Backup struct {
		Enabled bool `yaml:"enabled,omitempty"`
	} `yaml:"backup,omitempty"`
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "aidb", "config.yaml")
}

func loadUserConfig() (*UserConfig, error) {
	cfg := &UserConfig{}
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func saveUserConfig(cfg *UserConfig) error {
	path := getConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	userCfg, err := loadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// No args: show all config
	if len(args) == 0 {
		fmt.Println("# Current configuration")
		fmt.Println()
		fmt.Printf("db.path = %s\n", cfg.DBDir)
		fmt.Printf("backup.enabled = %v\n", userCfg.Backup.Enabled)
		fmt.Println()
		fmt.Printf("# Config file: %s\n", getConfigPath())
		return nil
	}

	key := args[0]

	// One arg: show specific key
	if len(args) == 1 {
		switch key {
		case "db.path":
			fmt.Println(cfg.DBDir)
		case "backup.enabled":
			fmt.Println(userCfg.Backup.Enabled)
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
		userCfg.Backup.Enabled = value == "true"
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	if err := saveUserConfig(userCfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	printSuccess(fmt.Sprintf("Set %s = %s", key, value))
	return nil
}
