package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// UserConfig is the persisted ~/.config/aidb/config.yaml.
// backup.enabled is intentionally absent: backup state is the launchd plist,
// managed by 'aidb backup enable/disable'.
type UserConfig struct {
	DB struct {
		Path string `yaml:"path,omitempty"`
	} `yaml:"db,omitempty"`
}

// UserConfigPath returns ~/.config/aidb/config.yaml
func UserConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "aidb", "config.yaml")
}

// LoadUserConfig reads the user config; a missing file yields an empty config
func LoadUserConfig() (*UserConfig, error) {
	cfg := &UserConfig{}
	data, err := os.ReadFile(UserConfigPath())
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

// SaveUserConfig writes the user config, creating parent directories
func SaveUserConfig(cfg *UserConfig) error {
	path := UserConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
