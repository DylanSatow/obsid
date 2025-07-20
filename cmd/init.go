/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize obsidian-cli configuration",
	Long: `Set up initial configuration for obsidian-cli with vault path and project directories.

This command creates a configuration file that tells obsidian-cli where your 
Obsidian vault is located and which project directories to monitor.

Examples:
  obsidian-cli init --vault ~/Obsidian/Main
  obsidian-cli init --vault ~/Obsidian/Main --projects ~/Projects,~/work`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringP("vault", "v", "", "path to Obsidian vault (required)")
	initCmd.Flags().StringSliceP("projects", "p", []string{}, "project directories to monitor")
	initCmd.MarkFlagRequired("vault")
}

func runInit(cmd *cobra.Command, args []string) error {
	vaultPath, _ := cmd.Flags().GetString("vault")
	projectDirs, _ := cmd.Flags().GetStringSlice("projects")

	// Expand home directory if needed
	if vaultPath[0] == '~' {
		home, _ := os.UserHomeDir()
		vaultPath = filepath.Join(home, vaultPath[1:])
	}

	// Validate vault path
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return fmt.Errorf("vault path does not exist: %s", vaultPath)
	}

	// Create config directory
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "obsidian-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	// Create default config
	config := map[string]interface{}{
		"vault": map[string]string{
			"path":             vaultPath,
			"daily_notes_dir":  "Daily Notes",
			"date_format":      "YYYY-MM-DD-dddd",
		},
		"projects": map[string]interface{}{
			"auto_discover": true,
			"directories":   projectDirs,
		},
		"git": map[string]interface{}{
			"include_diffs":        false,
			"max_commits":          10,
			"ignore_merge_commits": true,
		},
		"formatting": map[string]interface{}{
			"create_links":     true,
			"add_tags":         []string{"#programming"},
			"timestamp_format": "HH:mm",
		},
	}

	// Write config file
	configPath := filepath.Join(configDir, "config.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	fmt.Printf("✅ Configuration initialized at: %s\n", configPath)
	fmt.Printf("📁 Vault: %s\n", vaultPath)
	if len(projectDirs) > 0 {
		fmt.Printf("🚀 Project directories: %v\n", projectDirs)
	}
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  obsidian-cli log               # Log current project activity\n")
	fmt.Printf("  obsidian-cli log --git-summary # Include detailed git analysis\n")

	return nil
}
