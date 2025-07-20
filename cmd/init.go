/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize obsid configuration",
	Long: `Set up initial configuration for obsid with vault path and project directories.

This command runs in interactive mode by default, prompting you for all configuration
options. Use --non-interactive to provide all options via command-line flags.

Examples:
  obsid init                                              (interactive mode - recommended)
  obsid init --non-interactive --vault ~/Obsidian/Main   (non-interactive mode)`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringP("vault", "v", "", "path to Obsidian vault (required for non-interactive mode)")
	initCmd.Flags().StringSliceP("projects", "p", []string{}, "project directories to monitor")
	initCmd.Flags().BoolP("non-interactive", "n", false, "skip interactive prompts and use command-line flags")
	initCmd.Flags().StringP("daily-notes-dir", "", "Daily Notes", "daily notes directory name")
	initCmd.Flags().StringP("date-format", "", "YYYY-MM-DD-dddd", "date format for daily note filenames")
}

func runInit(cmd *cobra.Command, args []string) error {
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

	if nonInteractive {
		return runNonInteractiveInit(cmd)
	}

	return runInteractiveInit(cmd)
}

func runInteractiveInit(cmd *cobra.Command) error {
	// Create readline instance with basic tab completion
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "",
		AutoComplete: &BasicFileCompleter{},
	})
	if err != nil {
		return err
	}
	defer rl.Close()

	fmt.Println("Welcome to Obsidian CLI Setup!")
	fmt.Println("This interactive setup will help you configure obsid to work with your Obsidian vault.")
	fmt.Println()

	// Step 1: Get vault path
	vaultPath, err := promptForVaultPath(rl)
	if err != nil {
		return err
	}

	// Step 2: Auto-detect or configure daily notes
	dailyNotesDir, dateFormat, err := setupDailyNotes(vaultPath, rl)
	if err != nil {
		return err
	}

	// Step 3: Configure project directories
	projectDirs, err := promptForProjectDirectories(rl)
	if err != nil {
		return err
	}

	// Step 4: Configure git settings
	gitConfig, err := promptForGitSettings(rl)
	if err != nil {
		return err
	}

	// Step 5: Configure formatting options
	formatConfig, err := promptForFormattingSettings(rl)
	if err != nil {
		return err
	}

	// Create and save configuration
	return saveConfiguration(vaultPath, dailyNotesDir, dateFormat, projectDirs, gitConfig, formatConfig)
}

func runNonInteractiveInit(cmd *cobra.Command) error {
	vaultPath, _ := cmd.Flags().GetString("vault")
	projectDirs, _ := cmd.Flags().GetStringSlice("projects")
	dailyNotesDir, _ := cmd.Flags().GetString("daily-notes-dir")
	dateFormat, _ := cmd.Flags().GetString("date-format")

	// Validate required vault path for non-interactive mode
	if vaultPath == "" {
		return fmt.Errorf("vault path is required in non-interactive mode. Use --vault flag or run without --non-interactive")
	}

	// Expand home directory if needed
	if len(vaultPath) > 0 && vaultPath[0] == '~' {
		home, _ := os.UserHomeDir()
		vaultPath = filepath.Join(home, vaultPath[1:])
	}

	// Validate vault path
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return fmt.Errorf("vault path does not exist: %s", vaultPath)
	}

	// Use default configurations for non-interactive mode
	gitConfig := map[string]interface{}{
		"include_diffs":        false,
		"max_commits":          10,
		"ignore_merge_commits": true,
	}

	formatConfig := map[string]interface{}{
		"create_links":     true,
		"add_tags":         []string{"#programming"},
		"timestamp_format": "HH:mm",
	}

	return saveConfiguration(vaultPath, dailyNotesDir, dateFormat, projectDirs, gitConfig, formatConfig)
}

func promptForVaultPath(rl *readline.Instance) (string, error) {
	fmt.Println("Step 1: Obsidian Vault Location")
	fmt.Println("Please enter the path to your Obsidian vault.")
	fmt.Println("Examples: ~/Documents/MyVault, /Users/username/Obsidian/MainVault")
	fmt.Print("Vault path: ")

	rl.SetPrompt("Vault path: ")
	vaultPath, err := rl.Readline()
	if err != nil {
		return "", err
	}
	vaultPath = strings.TrimSpace(vaultPath)

	if vaultPath == "" {
		return "", fmt.Errorf("vault path cannot be empty")
	}

	// Expand home directory if needed
	if len(vaultPath) > 0 && vaultPath[0] == '~' {
		home, _ := os.UserHomeDir()
		if len(vaultPath) == 1 {
			vaultPath = home
		} else if vaultPath[1] == '/' {
			vaultPath = filepath.Join(home, vaultPath[2:])
		} else {
			vaultPath = filepath.Join(home, vaultPath[1:])
		}
	}

	// Validate vault path
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		fmt.Printf("Vault path does not exist: %s\n", vaultPath)
		fmt.Print("Would you like to try again? (y/N): ")
		rl.SetPrompt("Would you like to try again? (y/N): ")
		retry, _ := rl.Readline()
		if strings.ToLower(strings.TrimSpace(retry)) == "y" {
			return promptForVaultPath(rl)
		}
		return "", fmt.Errorf("vault path does not exist: %s", vaultPath)
	}

	fmt.Printf("Found vault at: %s\n\n", vaultPath)
	return vaultPath, nil
}

func setupDailyNotes(vaultPath string, rl *readline.Instance) (string, string, error) {
	fmt.Println("Step 2: Daily Notes Configuration")
	fmt.Println("Please configure your daily notes directory and date format.")
	return promptForDailyNoteConfig(vaultPath, "Daily Notes", "YYYY-MM-DD-dddd", rl)
}

func promptForProjectDirectories(rl *readline.Instance) ([]string, error) {
	fmt.Println("Step 3: Project Directories")
	fmt.Println("Enter directories containing your programming projects (optional).")
	fmt.Println("Examples: ~/Projects, ~/work, /Users/username/Development")
	fmt.Println("You can enter multiple directories separated by commas, or press Enter to skip.")
	fmt.Print("Project directories: ")

	rl.SetPrompt("Project directories: ")
	input, err := rl.Readline()
	if err != nil {
		return nil, err
	}
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("No project directories configured (you can add them later)\n")
		return []string{}, nil
	}

	// Split by comma and clean up
	dirs := strings.Split(input, ",")
	var projectDirs []string
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir != "" {
			// Expand home directory if needed
			if len(dir) > 0 && dir[0] == '~' {
				home, _ := os.UserHomeDir()
				if len(dir) == 1 {
					dir = home
				} else if dir[1] == '/' {
					dir = filepath.Join(home, dir[2:])
				} else {
					dir = filepath.Join(home, dir[1:])
				}
			}
			projectDirs = append(projectDirs, dir)
		}
	}

	fmt.Printf("Configured %d project directories\n\n", len(projectDirs))
	return projectDirs, nil
}

func promptForGitSettings(rl *readline.Instance) (map[string]interface{}, error) {
	fmt.Println("Step 4: Git Analysis Settings")

	// Max commits
	fmt.Print("Maximum commits to analyze (default: 10): ")
	rl.SetPrompt("Maximum commits to analyze (default: 10): ")
	maxCommitsStr, _ := rl.Readline()
	maxCommitsStr = strings.TrimSpace(maxCommitsStr)
	maxCommits := 10
	if maxCommitsStr != "" {
		if parsed, err := strconv.Atoi(maxCommitsStr); err == nil && parsed > 0 {
			maxCommits = parsed
		}
	}

	// Include diffs
	fmt.Print("Include file diffs in analysis? (y/N): ")
	rl.SetPrompt("Include file diffs in analysis? (y/N): ")
	includeDiffsStr, _ := rl.Readline()
	includeDiffs := strings.ToLower(strings.TrimSpace(includeDiffsStr)) == "y"

	// Ignore merge commits
	fmt.Print("Ignore merge commits? (Y/n): ")
	rl.SetPrompt("Ignore merge commits? (Y/n): ")
	ignoreMergeStr, _ := rl.Readline()
	ignoreMergeStr = strings.TrimSpace(ignoreMergeStr)
	ignoreMerge := ignoreMergeStr == "" || strings.ToLower(ignoreMergeStr) == "y"

	fmt.Println("Git settings configured\n")

	return map[string]interface{}{
		"include_diffs":        includeDiffs,
		"max_commits":          maxCommits,
		"ignore_merge_commits": ignoreMerge,
	}, nil
}

func promptForFormattingSettings(rl *readline.Instance) (map[string]interface{}, error) {
	fmt.Println("Step 5: Formatting Options")

	// Create links
	fmt.Print("Create Obsidian links for file names? (Y/n): ")
	rl.SetPrompt("Create Obsidian links for file names? (Y/n): ")
	createLinksStr, _ := rl.Readline()
	createLinksStr = strings.TrimSpace(createLinksStr)
	createLinks := createLinksStr == "" || strings.ToLower(createLinksStr) == "y"

	// Tags
	fmt.Print("Default tags to add (comma-separated, default: #programming): ")
	rl.SetPrompt("Default tags to add (comma-separated, default: #programming): ")
	tagsStr, _ := rl.Readline()
	tagsStr = strings.TrimSpace(tagsStr)
	tags := []string{"#programming"}
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tag = strings.TrimSpace(tag)
			if !strings.HasPrefix(tag, "#") {
				tag = "#" + tag
			}
			tags[i] = tag
		}
	}

	// Timestamp format
	fmt.Print("Timestamp format (default: HH:mm): ")
	rl.SetPrompt("Timestamp format (default: HH:mm): ")
	timestampStr, _ := rl.Readline()
	timestampStr = strings.TrimSpace(timestampStr)
	if timestampStr == "" {
		timestampStr = "HH:mm"
	}

	fmt.Println("Formatting settings configured\n")

	return map[string]interface{}{
		"create_links":     createLinks,
		"add_tags":         tags,
		"timestamp_format": timestampStr,
	}, nil
}

func saveConfiguration(vaultPath, dailyNotesDir, dateFormat string, projectDirs []string, gitConfig, formatConfig map[string]interface{}) error {
	// Create config directory
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "obsid")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	// Create configuration
	config := map[string]interface{}{
		"vault": map[string]string{
			"path":            vaultPath,
			"daily_notes_dir": dailyNotesDir,
			"date_format":     dateFormat,
		},
		"projects": map[string]interface{}{
			"auto_discover": true,
			"directories":   projectDirs,
		},
		"git":        gitConfig,
		"formatting": formatConfig,
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

	// Success message
	fmt.Println("Configuration Complete!")
	fmt.Printf("Configuration saved to: %s\n\n", configPath)
	fmt.Println("Summary:")
	fmt.Printf("   Vault: %s\n", vaultPath)
	fmt.Printf("   Daily notes directory: %s\n", dailyNotesDir)
	fmt.Printf("   Date format: %s\n", dateFormat)
	if len(projectDirs) > 0 {
		fmt.Printf("   Project directories: %v\n", projectDirs)
	}
	fmt.Println()
	fmt.Println("Ready to use: obsid log")
	fmt.Println("   Run this command from any git repository to log your activity!")

	return nil
}

func promptForDailyNoteConfig(vaultPath, currentDailyNotesDir, currentDateFormat string, rl *readline.Instance) (string, string, error) {

	fmt.Println("\nInteractive Daily Note Configuration")
	fmt.Println("This will help configure obsid to work with your specific daily note setup.")

	// Ask for daily notes directory
	fmt.Printf("\nDaily notes directory (current: %s): ", currentDailyNotesDir)
	fmt.Print("Enter the folder name where your daily notes are stored (press Enter for default): ")

	rl.SetPrompt("Enter the folder name where your daily notes are stored (press Enter for default): ")
	dailyNotesDir, err := rl.Readline()
	if err != nil {
		return "", "", err
	}
	dailyNotesDir = strings.TrimSpace(dailyNotesDir)
	if dailyNotesDir == "" {
		dailyNotesDir = currentDailyNotesDir
	}

	// Show common date format examples
	fmt.Println("\nCommon daily note date formats:")
	fmt.Println("  1. YYYY-MM-DD-dddd          → 2025-07-20-Sunday")
	fmt.Println("  2. YYYY-MM-DD               → 2025-07-20")
	fmt.Println("  3. DD-MM-YYYY               → 20-07-2025")
	fmt.Println("  4. MM-DD-YYYY               → 07-20-2025")
	fmt.Println("  5. MM-DD-YY                 → 07-20-25")
	fmt.Println("  6. YYYY/MM/DD               → 2025/07/20")
	fmt.Println("  7. MMMM DD, YYYY            → July 20, 2025")
	fmt.Println("  8. DD MMMM YYYY             → 20 July 2025")
	fmt.Println("  9. YYYY-MM-DD dddd          → 2025-07-20 Sunday")
	fmt.Println(" 10. YY-MM-DD                 → 25-07-20")

	fmt.Printf("\nCurrent format: %s\n", currentDateFormat)
	fmt.Print("Enter your date format (press Enter to keep current, or type a number 1-10 for common formats): ")

	rl.SetPrompt("Enter your date format (press Enter to keep current, or type a number 1-10 for common formats): ")
	dateFormatInput, err := rl.Readline()
	if err != nil {
		return "", "", err
	}
	dateFormatInput = strings.TrimSpace(dateFormatInput)

	var dateFormat string
	switch dateFormatInput {
	case "":
		dateFormat = currentDateFormat
	case "1":
		dateFormat = "YYYY-MM-DD-dddd"
	case "2":
		dateFormat = "YYYY-MM-DD"
	case "3":
		dateFormat = "DD-MM-YYYY"
	case "4":
		dateFormat = "MM-DD-YYYY"
	case "5":
		dateFormat = "MM-DD-YY"
	case "6":
		dateFormat = "YYYY/MM/DD"
	case "7":
		dateFormat = "MMMM DD, YYYY"
	case "8":
		dateFormat = "DD MMMM YYYY"
	case "9":
		dateFormat = "YYYY-MM-DD dddd"
	case "10":
		dateFormat = "YY-MM-DD"
	default:
		dateFormat = dateFormatInput
	}

	// Validate the configuration
	fmt.Printf("\nConfiguration Summary:\n")
	fmt.Printf("   Daily notes directory: %s\n", dailyNotesDir)
	fmt.Printf("   Date format: %s\n", dateFormat)

	// Show what today's note would be named
	exampleName := formatDateExample(dateFormat)
	fmt.Printf("   Today's note would be: %s/%s.md\n", dailyNotesDir, exampleName)

	fmt.Print("\nDoes this look correct? (y/N): ")
	rl.SetPrompt("Does this look correct? (y/N): ")
	confirm, err := rl.Readline()
	if err != nil {
		return "", "", err
	}

	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("Configuration cancelled. Please run with --interactive again.")
		return "", "", fmt.Errorf("configuration cancelled by user")
	}

	return dailyNotesDir, dateFormat, nil
}



func formatDateExample(format string) string {
	// Convert our YYYY-MM-DD format to what it would look like today
	// This is a simplified example - in real usage we'd use the actual date formatting
	switch format {
	case "YYYY-MM-DD-dddd":
		return "2025-07-20-Sunday"
	case "YYYY-MM-DD":
		return "2025-07-20"
	case "DD-MM-YYYY":
		return "20-07-2025"
	case "MM-DD-YYYY":
		return "07-20-2025"
	case "MM-DD-YY":
		return "07-20-25"
	case "YYYY/MM/DD":
		return "2025/07/20"
	case "MMMM DD, YYYY":
		return "July 20, 2025"
	case "DD MMMM YYYY":
		return "20 July 2025"
	case "YYYY-MM-DD dddd":
		return "2025-07-20 Sunday"
	case "YY-MM-DD":
		return "25-07-20"
	default:
		return format // Return the format itself as an example
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// BasicFileCompleter provides simple file completion
type BasicFileCompleter struct{}

// Do implements the AutoCompleter interface with basic file completion
func (c *BasicFileCompleter) Do(line []rune, pos int) ([][]rune, int) {
	if pos == 0 {
		return nil, 0
	}
	
	// Note: we work directly with the rune slice
	
	// Find the start of the current word
	start := pos - 1
	for start >= 0 && line[start] != ' ' && line[start] != '\t' {
		start--
	}
	start++
	
	// Extract the current path being typed
	currentPath := string(line[start:pos])
	
	// Get directory to search and file prefix
	var searchDir, filePrefix string
	
	if currentPath == "" {
		searchDir = "."
		filePrefix = ""
	} else if strings.Contains(currentPath, "/") {
		searchDir = filepath.Dir(currentPath)
		filePrefix = filepath.Base(currentPath)
		
		// Handle special cases
		if searchDir == "." && strings.HasPrefix(currentPath, "./") {
			searchDir = "."
		} else if strings.HasPrefix(currentPath, "~/") {
			home, _ := os.UserHomeDir()
			if searchDir == "~" {
				searchDir = home
			} else {
				searchDir = strings.Replace(searchDir, "~", home, 1)
			}
		}
	} else {
		searchDir = "."
		filePrefix = currentPath
	}
	
	// Read directory contents
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil, len(currentPath)
	}
	
	// Find matching entries and return completions
	var completions [][]rune
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, filePrefix) {
			// Calculate the suffix to complete
			suffix := name[len(filePrefix):]
			if entry.IsDir() {
				suffix += "/"
			}
			if suffix != "" {
				completions = append(completions, []rune(suffix))
			}
		}
	}
	
	return completions, len(currentPath)
}


