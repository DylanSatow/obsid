package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var GlobalConfig *Config

var viperInstance *viper.Viper

func LoadConfig() error {
	viperInstance = viper.New()

	// Set config name and paths
	viperInstance.SetConfigName("config")
	viperInstance.SetConfigType("yaml")

	// Add config paths
	home, _ := os.UserHomeDir()
	viperInstance.AddConfigPath(filepath.Join(home, ".config", "obsidian-cli"))
	viperInstance.AddConfigPath(".")

	// Enable environment variable reading
	viperInstance.SetEnvPrefix("OBSIDIAN_CLI")
	viperInstance.AutomaticEnv()

	// Set defaults before reading config
	setDefaults(viperInstance)

	// Read config file
	if err := viperInstance.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal into struct
	GlobalConfig = &Config{}
	if err := viperInstance.Unmarshal(GlobalConfig); err != nil {
		return err
	}
	
	return nil
}

// GetViperValue returns the actual value from viper, bypassing GlobalConfig if needed
func GetViperValue(key string) string {
	if viperInstance != nil {
		return viperInstance.GetString(key)
	}
	return ""
}

func setDefaults(v *viper.Viper) {
	// Set defaults for all configuration values
	v.SetDefault("vault.path", "")
	v.SetDefault("vault.daily_notes_dir", "Daily Notes")
	v.SetDefault("vault.date_format", "YYYY-MM-DD-dddd")
	v.SetDefault("projects.auto_discover", true)
	v.SetDefault("projects.directories", []string{})
	v.SetDefault("git.include_diffs", false)
	v.SetDefault("git.max_commits", 10)
	v.SetDefault("git.ignore_merge_commits", true)
	v.SetDefault("formatting.create_links", true)
	v.SetDefault("formatting.add_tags", []string{"#programming"})
	v.SetDefault("formatting.timestamp_format", "HH:mm")
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "obsidian-cli", "config.yaml")
}

func ConfigExists() bool {
	_, err := os.Stat(GetConfigPath())
	return err == nil
}