/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/DylanSatow/obsidian-cli/pkg/config"
	"github.com/DylanSatow/obsidian-cli/pkg/git"
	"github.com/DylanSatow/obsidian-cli/pkg/obsidian"
	"github.com/DylanSatow/obsidian-cli/pkg/utils"
	"github.com/spf13/cobra"
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Log current project activity to daily note",
	Long: `Analyze current project and log git commits and changes to today's daily note.

This command finds the current git repository, analyzes recent commits and changes,
then formats and appends this information to your Obsidian daily note under a 
"Projects" section.

Examples:
  obsidian-cli log                                    # Log last hour of activity
  obsidian-cli log --git-summary                     # Include detailed git analysis  
  obsidian-cli log --timeframe 2h                    # Log last 2 hours
  obsidian-cli log --timeframe today                 # Log all activity today
  obsidian-cli log --project "My Custom Project"     # Override project name
  obsidian-cli log --create-note                     # Create daily note if missing`,
	RunE: runLog,
}

func init() {
	rootCmd.AddCommand(logCmd)

	logCmd.Flags().BoolP("git-summary", "g", false, "include detailed git analysis")
	logCmd.Flags().StringP("timeframe", "t", "1h", "timeframe for analysis (e.g., '2h', 'today')")
	logCmd.Flags().StringP("project", "p", "", "override project name")
	logCmd.Flags().BoolP("create-note", "c", false, "create daily note if it doesn't exist")
}

func runLog(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current directory: %w", err)
	}

	// Find git repository
	repo, err := git.FindRepository(cwd)
	if err != nil {
		return fmt.Errorf("could not find git repository: %w", err)
	}

	// Parse timeframe
	timeframe, _ := cmd.Flags().GetString("timeframe")
	since, err := utils.ParseTimeframe(timeframe)
	if err != nil {
		return fmt.Errorf("invalid timeframe: %w", err)
	}

	// Get project name (use flag override or repository name)
	projectName, _ := cmd.Flags().GetString("project")
	if projectName == "" {
		projectName = repo.Name
	}

	// Get commits
	commits, err := repo.GetCommits(since, config.GlobalConfig.Git.MaxCommits)
	if err != nil {
		return fmt.Errorf("could not get commits: %w", err)
	}

	// Get changed files if git-summary is requested
	var files []string
	gitSummary, _ := cmd.Flags().GetBool("git-summary")
	if gitSummary {
		files, err = repo.GetChangedFiles(since)
		if err != nil {
			fmt.Printf("Warning: could not get changed files: %v\n", err)
		}
	}

	// Create vault instance - use viper values if GlobalConfig is empty
	vaultPath := config.GlobalConfig.Vault.Path
	dailyNotesDir := config.GlobalConfig.Vault.DailyNotesDir
	dateFormat := config.GlobalConfig.Vault.DateFormat
	
	// Fallback to viper if GlobalConfig is empty
	if vaultPath == "" {
		vaultPath = config.GetViperValue("vault.path")
	}
	if dailyNotesDir == "" {
		dailyNotesDir = config.GetViperValue("vault.daily_notes_dir")
	}
	if dateFormat == "" {
		dateFormat = config.GetViperValue("vault.date_format")
	}
	
	vault := obsidian.NewVault(vaultPath, dailyNotesDir, dateFormat)

	// Validate vault exists
	if !vault.Exists() {
		return fmt.Errorf("vault not found at: %s", vault.Path)
	}

	// Check if daily note exists and handle creation
	today := time.Now()
	createNote, _ := cmd.Flags().GetBool("create-note")
	
	// First try to find an existing daily note using any format
	_, exists := vault.FindExistingDailyNote(today)
	
	if !exists {
		if !createNote {
			return fmt.Errorf("daily note does not exist for %s\n\nUse --create-note flag to create it automatically:\n  obsidian-cli log --create-note", today.Format("Monday, January 2, 2006"))
		}
		
		if err := vault.CreateDailyNote(today); err != nil {
			return fmt.Errorf("could not create daily note: %w", err)
		}
		fmt.Printf("📄 Created new daily note for %s\n", today.Format("Monday, January 2, 2006"))
	} else {
		// Update vault's date format to match the existing note format
		detectedFormat, err := vault.DetectDateFormat()
		if err == nil && detectedFormat != "" && detectedFormat != vault.DateFormat {
			vault.DateFormat = detectedFormat
		}
	}

	// Format project entry
	timeRange := utils.FormatTimeRange(since)
	content := obsidian.FormatProjectEntry(repo, commits, files, timeRange)

	// Append to daily note
	if err := vault.AppendProjectEntry(today, projectName, content); err != nil {
		return fmt.Errorf("could not append to daily note: %w", err)
	}

	// Success message
	fmt.Printf("✅ Logged activity for %s\n", projectName)
	fmt.Printf("📝 Daily note: %s\n", vault.GetDailyNotePath(today))
	fmt.Printf("⏰ Time range: %s\n", timeRange)
	if len(commits) > 0 {
		fmt.Printf("📦 Commits: %d\n", len(commits))
	}
	if len(files) > 0 {
		fmt.Printf("📁 Files changed: %d\n", len(files))
	}

	return nil
}
