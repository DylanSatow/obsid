/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/DylanSatow/obsidian-cli/pkg/config"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "obsidian-cli",
	Short: "A CLI tool for logging programming projects to Obsidian daily notes",
	Long: `Obsidian CLI automates the logging of programming project activities 
into your Obsidian daily notes. Track git commits, changes, and project 
progress with intelligent formatting and seamless integration.

Examples:
  obsidian-cli init --vault ~/Obsidian/Main
  obsidian-cli log
  obsidian-cli log --git-summary --timeframe 2h`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip config loading for init command
		if cmd.Name() == "init" {
			return
		}
		
		if err := config.LoadConfig(); err != nil {
			if !config.ConfigExists() {
				fmt.Println("No configuration found. Run 'obsidian-cli init' to set up.")
				os.Exit(1)
			}
			fmt.Printf("Warning: Could not load config: %v\n", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("vault", "v", "", "path to Obsidian vault")
	rootCmd.PersistentFlags().BoolP("verbose", "", false, "enable verbose output")
}


