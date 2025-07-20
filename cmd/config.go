/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/DylanSatow/obsidian-cli/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	Long: `Display the current obsidian-cli configuration.

This command shows the loaded configuration including vault path, 
project directories, and formatting settings.

Examples:
  obsidian-cli config       # Show current configuration`,
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	fmt.Printf("Configuration file: %s\n\n", config.GetConfigPath())
	
	// Show the actual loaded configuration
	if config.GlobalConfig == nil {
		return fmt.Errorf("configuration not loaded")
	}
	
	// Convert to YAML for nice display
	data, err := yaml.Marshal(config.GlobalConfig)
	if err != nil {
		return fmt.Errorf("could not format configuration: %w", err)
	}

	fmt.Println("Current configuration:")
	fmt.Println(string(data))
	
	return nil
}
