/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DylanSatow/obsid/pkg/config"
	"github.com/DylanSatow/obsid/pkg/git"
	"github.com/DylanSatow/obsid/pkg/obsidian"
	"github.com/DylanSatow/obsid/pkg/utils"
	"github.com/spf13/cobra"
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log [path]",
	Short: "Log project activity to daily note",
	Long: `Analyze project(s) and log git commits and changes to today's daily note.

When a path is provided, logs only that specific repository.
When no path is provided, recursively discovers and logs all git repositories 
in your configured projects directories.

Examples:
  obsid log                                    # Log all repos in projects directories
  obsid log .                                  # Log current directory repo
  obsid log /path/to/repo                      # Log specific repo
  obsid log --git-summary                     # Include detailed git analysis  
  obsid log --timeframe 2h                    # Log last 2 hours
  obsid log --timeframe today                 # Log all activity today
  obsid log --project "My Custom Project"     # Override project name
  obsid log --create-note                     # Create daily note if missing`,
	RunE: runLog,
}

func init() {
	rootCmd.AddCommand(logCmd)

	logCmd.Flags().BoolP("git-summary", "g", false, "include detailed git analysis")
	logCmd.Flags().StringP("timeframe", "t", "1h", "timeframe for analysis (e.g., '2h', 'today')")
	logCmd.Flags().StringP("project", "p", "", "override project name")
	logCmd.Flags().BoolP("create-note", "c", false, "create daily note if it doesn't exist")
}

func discoverGitRepositories(directories []string) ([]*git.Repository, error) {
	var repos []*git.Repository
	
	for _, dir := range directories {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip inaccessible paths
			}
			
			if info.IsDir() && info.Name() == ".git" {
				repoPath := filepath.Dir(path)
				repo, err := git.FindRepository(repoPath)
				if err == nil {
					repos = append(repos, repo)
				}
				return filepath.SkipDir // Don't go deeper into .git directory
			}
			
			return nil
		})
		
		if err != nil {
			fmt.Printf("Warning: could not scan directory %s: %v\n", dir, err)
		}
	}
	
	return repos, nil
}

func logSingleRepository(repo *git.Repository, cmd *cobra.Command) error {
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

	// Skip if no activity
	if len(commits) == 0 {
		return nil
	}

	// Get changed files if git-summary is requested
	var files []string
	gitSummary, _ := cmd.Flags().GetBool("git-summary")
	if gitSummary {
		files, err = repo.GetChangedFiles(since)
		if err != nil {
			fmt.Printf("Warning: could not get changed files for %s: %v\n", repo.Name, err)
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
			return fmt.Errorf("daily note does not exist for %s\n\nUse --create-note flag to create it automatically:\n  obsid log --create-note", today.Format("Monday, January 2, 2006"))
		}
		
		if err := vault.CreateDailyNote(today); err != nil {
			return fmt.Errorf("could not create daily note: %w", err)
		}
		fmt.Printf("Created new daily note for %s\n", today.Format("Monday, January 2, 2006"))
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
	fmt.Printf("Logged activity for %s (commits: %d", projectName, len(commits))
	if len(files) > 0 {
		fmt.Printf(", files: %d", len(files))
	}
	fmt.Printf(")\n")

	return nil
}

func runLog(cmd *cobra.Command, args []string) error {
	var repos []*git.Repository
	
	if len(args) > 0 {
		// Path provided - log specific repository
		targetPath := args[0]
		if targetPath == "." {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("could not get current directory: %w", err)
			}
			targetPath = cwd
		}
		
		repo, err := git.FindRepository(targetPath)
		if err != nil {
			return fmt.Errorf("could not find git repository at %s: %w", targetPath, err)
		}
		repos = append(repos, repo)
	} else {
		// No path provided - discover all repositories in projects directories
		projectDirs := config.GlobalConfig.Projects.Directories
		if len(projectDirs) == 0 {
			// Fallback to default projects directory
			home, _ := os.UserHomeDir()
			projectDirs = []string{filepath.Join(home, "projects")}
		}
		
		discoveredRepos, err := discoverGitRepositories(projectDirs)
		if err != nil {
			return fmt.Errorf("could not discover repositories: %w", err)
		}
		repos = discoveredRepos
	}
	
	if len(repos) == 0 {
		return fmt.Errorf("no git repositories found")
	}
	
	// Log each repository
	loggedCount := 0
	for _, repo := range repos {
		if err := logSingleRepository(repo, cmd); err != nil {
			fmt.Printf("Error logging %s: %v\n", repo.Name, err)
			continue
		}
		loggedCount++
	}
	
	if loggedCount == 0 {
		return fmt.Errorf("no repositories had activity to log")
	}
	
	fmt.Printf("\nLogged %d of %d repositories\n", loggedCount, len(repos))
	return nil
}
