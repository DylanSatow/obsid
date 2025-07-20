/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

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
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to Obsidian CLI Setup!")
	fmt.Println("This interactive setup will help you configure obsid to work with your Obsidian vault.")
	fmt.Println()

	// Step 1: Get vault path
	vaultPath, err := promptForVaultPath(reader)
	if err != nil {
		return err
	}

	// Step 2: Auto-detect or configure daily notes
	dailyNotesDir, dateFormat, err := setupDailyNotes(vaultPath, reader)
	if err != nil {
		return err
	}

	// Step 3: Configure project directories
	projectDirs, err := promptForProjectDirectories(reader)
	if err != nil {
		return err
	}

	// Step 4: Configure git settings
	gitConfig, err := promptForGitSettings(reader)
	if err != nil {
		return err
	}

	// Step 5: Configure formatting options
	formatConfig, err := promptForFormattingSettings(reader)
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

func promptForVaultPath(reader *bufio.Reader) (string, error) {
	fmt.Println("Step 1: Obsidian Vault Location")
	fmt.Println("Please enter the path to your Obsidian vault.")
	fmt.Println("Examples: ~/Documents/MyVault, /Users/username/Obsidian/MainVault")
	fmt.Print("Vault path: ")

	vaultPath, err := reader.ReadString('\n')
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
		retry, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(retry)) == "y" {
			return promptForVaultPath(reader)
		}
		return "", fmt.Errorf("vault path does not exist: %s", vaultPath)
	}

	fmt.Printf("Found vault at: %s\n\n", vaultPath)
	return vaultPath, nil
}

func setupDailyNotes(vaultPath string, reader *bufio.Reader) (string, string, error) {
	fmt.Println("Step 2: Daily Notes Configuration")
	fmt.Println("Scanning your vault for existing daily notes...")

	// First scan for existing daily notes in any directory
	suggestions := scanVaultForDailyNotes(vaultPath)
	
	if len(suggestions) > 0 {
		fmt.Printf("Found %d existing daily notes in your vault. Here are some patterns:\n", len(suggestions))
		for i, suggestion := range suggestions[:min(5, len(suggestions))] {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
		
		// Try to auto-detect the most common directory and format
		detectedDir, detectedFormat := detectDailyNotesConfig(suggestions)
		
		if detectedDir != "" && detectedFormat != "" {
			fmt.Printf("\nDetected configuration:\n")
			fmt.Printf("  Directory: %s\n", detectedDir)
			fmt.Printf("  Format: %s\n", detectedFormat)
			fmt.Print("Use detected configuration? (Y/n): ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(response)
			if response == "" || strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
				return detectedDir, detectedFormat, nil
			}
		}
	} else {
		fmt.Println("No existing daily notes found.")
	}

	fmt.Println("Let's configure your daily notes setup.")
	return promptForDailyNoteConfig(vaultPath, "Daily Notes", "YYYY-MM-DD-dddd")
}

func promptForProjectDirectories(reader *bufio.Reader) ([]string, error) {
	fmt.Println("Step 3: Project Directories")
	fmt.Println("Enter directories containing your programming projects (optional).")
	fmt.Println("Examples: ~/Projects, ~/work, /Users/username/Development")
	fmt.Println("You can enter multiple directories separated by commas, or press Enter to skip.")
	fmt.Print("Project directories: ")

	input, err := reader.ReadString('\n')
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

func promptForGitSettings(reader *bufio.Reader) (map[string]interface{}, error) {
	fmt.Println("Step 4: Git Analysis Settings")

	// Max commits
	fmt.Print("Maximum commits to analyze (default: 10): ")
	maxCommitsStr, _ := reader.ReadString('\n')
	maxCommitsStr = strings.TrimSpace(maxCommitsStr)
	maxCommits := 10
	if maxCommitsStr != "" {
		if parsed, err := strconv.Atoi(maxCommitsStr); err == nil && parsed > 0 {
			maxCommits = parsed
		}
	}

	// Include diffs
	fmt.Print("Include file diffs in analysis? (y/N): ")
	includeDiffsStr, _ := reader.ReadString('\n')
	includeDiffs := strings.ToLower(strings.TrimSpace(includeDiffsStr)) == "y"

	// Ignore merge commits
	fmt.Print("Ignore merge commits? (Y/n): ")
	ignoreMergeStr, _ := reader.ReadString('\n')
	ignoreMergeStr = strings.TrimSpace(ignoreMergeStr)
	ignoreMerge := ignoreMergeStr == "" || strings.ToLower(ignoreMergeStr) == "y"

	fmt.Println("Git settings configured\n")

	return map[string]interface{}{
		"include_diffs":        includeDiffs,
		"max_commits":          maxCommits,
		"ignore_merge_commits": ignoreMerge,
	}, nil
}

func promptForFormattingSettings(reader *bufio.Reader) (map[string]interface{}, error) {
	fmt.Println("Step 5: Formatting Options")

	// Create links
	fmt.Print("Create Obsidian links for file names? (Y/n): ")
	createLinksStr, _ := reader.ReadString('\n')
	createLinksStr = strings.TrimSpace(createLinksStr)
	createLinks := createLinksStr == "" || strings.ToLower(createLinksStr) == "y"

	// Tags
	fmt.Print("Default tags to add (comma-separated, default: #programming): ")
	tagsStr, _ := reader.ReadString('\n')
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
	timestampStr, _ := reader.ReadString('\n')
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

func promptForDailyNoteConfig(vaultPath, currentDailyNotesDir, currentDateFormat string) (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nInteractive Daily Note Configuration")
	fmt.Println("This will help configure obsid to work with your specific daily note setup.")

	// Ask for daily notes directory
	fmt.Printf("\nDaily notes directory (current: %s): ", currentDailyNotesDir)
	fmt.Print("Enter the folder name where your daily notes are stored (press Enter for default): ")

	dailyNotesDir, err := reader.ReadString('\n')
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

	dateFormatInput, err := reader.ReadString('\n')
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
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		fmt.Println("Configuration cancelled. Please run with --interactive again.")
		return "", "", fmt.Errorf("configuration cancelled by user")
	}

	return dailyNotesDir, dateFormat, nil
}

func scanVaultForDailyNotes(vaultPath string) []string {
	var suggestions []string

	// Look for markdown files that might be daily notes
	err := filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			// Check if filename looks like a date
			relPath, _ := filepath.Rel(vaultPath, path)
			if looksLikeDailyNote(info.Name()) {
				suggestions = append(suggestions, relPath)
			}
		}

		return nil
	})

	if err != nil {
		return []string{}
	}

	return suggestions
}

func looksLikeDailyNote(filename string) bool {
	name := strings.TrimSuffix(filename, ".md")

	// Check for common date patterns
	patterns := []string{
		"2025", "2024", "2023", // Years
		"01-", "02-", "03-", "04-", "05-", "06-", // Months
		"07-", "08-", "09-", "10-", "11-", "12-",
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
		"Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
		"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday",
	}

	for _, pattern := range patterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	return false
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

// detectDailyNotesConfig analyzes the file paths to detect the most common directory and format
func detectDailyNotesConfig(suggestions []string) (string, string) {
	dirCounts := make(map[string]int)
	formatCounts := make(map[string]int)
	
	for _, suggestion := range suggestions {
		// Extract directory
		dir := filepath.Dir(suggestion)
		if dir == "." {
			dir = "" // Root of vault
		}
		dirCounts[dir]++
		
		// Extract filename and try to detect format
		filename := filepath.Base(suggestion)
		filename = strings.TrimSuffix(filename, ".md")
		
		// Use the same detection logic as the vault package
		detectedFormat := detectDateFormatFromSingleFile(filename)
		if detectedFormat != "" {
			formatCounts[detectedFormat]++
		}
	}
	
	// Find most common directory
	mostCommonDir := ""
	maxDirCount := 0
	for dir, count := range dirCounts {
		if count > maxDirCount {
			maxDirCount = count
			mostCommonDir = dir
		}
	}
	
	// Find most common format
	mostCommonFormat := ""
	maxFormatCount := 0
	for format, count := range formatCounts {
		if count > maxFormatCount {
			maxFormatCount = count
			mostCommonFormat = format
		}
	}
	
	// Use the detected directory, or default if empty/root
	if mostCommonDir == "" || mostCommonDir == "." {
		mostCommonDir = "Daily Notes"
	}
	
	return mostCommonDir, mostCommonFormat
}

// detectDateFormatFromSingleFile detects the date format from a single filename
func detectDateFormatFromSingleFile(filename string) string {
	// Define patterns and their corresponding formats
	patterns := map[string]string{
		`^\d{4}-\d{2}-\d{2}-\w+$`:     "YYYY-MM-DD-dddd",    // 2025-07-20-Sunday
		`^\d{4}-\d{2}-\d{2} \w+$`:     "YYYY-MM-DD dddd",    // 2025-07-20 Sunday
		`^\d{4}-\d{2}-\d{2}$`:         "YYYY-MM-DD",         // 2025-07-20
		`^\d{2}-\d{2}-\d{4}$`:         "DD-MM-YYYY",         // 20-07-2025
		`^\d{1,2}-\d{1,2}-\d{4}$`:     "MM-DD-YYYY",         // 7-20-2025 or 07-20-2025
		`^\d{1,2}-\d{1,2}-\d{2}$`:     "MM-DD-YY",           // 7-20-25 or 07-20-25
		`^\d{4}/\d{2}/\d{2}$`:         "YYYY/MM/DD",         // 2025/07/20
		`^[A-Z][a-z]+ \d{1,2}, \d{4}$`: "MMMM DD, YYYY",    // July 20, 2025
		`^\d{1,2} [A-Z][a-z]+ \d{4}$`: "DD MMMM YYYY",      // 20 July 2025
		`^\d{2}-\d{2}-\d{2}$`:         "YY-MM-DD",           // 25-07-20
	}
	
	for pattern, format := range patterns {
		if matched, _ := regexp.MatchString(pattern, filename); matched {
			return format
		}
	}
	
	return ""
}
