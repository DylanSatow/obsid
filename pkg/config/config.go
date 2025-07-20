package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var GlobalConfig *Config

func LoadConfig() error {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add config paths
	home, _ := os.UserHomeDir()
	v.AddConfigPath(filepath.Join(home, ".config", "obsidian-cli"))
	v.AddConfigPath(".")

	// Enable environment variable reading
	v.SetEnvPrefix("OBSIDIAN_CLI")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal into struct
	GlobalConfig = &Config{}
	if err := v.Unmarshal(GlobalConfig); err != nil {
		return err
	}
	
	// Apply defaults if values are empty
	if GlobalConfig.Vault.DailyNotesDir == "" {
		GlobalConfig.Vault.DailyNotesDir = "Daily Notes"
	}
	if GlobalConfig.Vault.DateFormat == "" {
		GlobalConfig.Vault.DateFormat = "YYYY-MM-DD-dddd"
	}
	if GlobalConfig.Git.MaxCommits == 0 {
		GlobalConfig.Git.MaxCommits = 10
	}
	if GlobalConfig.Format.TimestampFormat == "" {
		GlobalConfig.Format.TimestampFormat = "HH:mm"
	}
	if len(GlobalConfig.Format.AddTags) == 0 {
		GlobalConfig.Format.AddTags = []string{"#programming"}
	}
	
	return nil
}

func setDefaults(v *viper.Viper) {
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