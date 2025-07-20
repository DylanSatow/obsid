/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DylanSatow/obsidian-cli/pkg/obsidian"
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
  obsidian-cli init --vault ~/Obsidian/Main --projects ~/Projects,~/work
  obsidian-cli init --vault ~/Obsidian/Main --interactive`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringP("vault", "v", "", "path to Obsidian vault (required)")
	initCmd.Flags().StringSliceP("projects", "p", []string{}, "project directories to monitor")
	initCmd.Flags().BoolP("interactive", "i", false, "interactive setup with prompts for daily note configuration")
	initCmd.Flags().StringP("daily-notes-dir", "", "Daily Notes", "daily notes directory name")
	initCmd.Flags().StringP("date-format", "", "YYYY-MM-DD-dddd", "date format for daily note filenames")
	initCmd.MarkFlagRequired("vault")
}

func runInit(cmd *cobra.Command, args []string) error {
	vaultPath, _ := cmd.Flags().GetString("vault")
	projectDirs, _ := cmd.Flags().GetStringSlice("projects")
	interactive, _ := cmd.Flags().GetBool("interactive")
	dailyNotesDir, _ := cmd.Flags().GetString("daily-notes-dir")
	dateFormat, _ := cmd.Flags().GetString("date-format")

	// Expand home directory if needed
	if vaultPath[0] == '~' {
		home, _ := os.UserHomeDir()
		vaultPath = filepath.Join(home, vaultPath[1:])
	}

	// Validate vault path
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return fmt.Errorf("vault path does not exist: %s", vaultPath)
	}

	// Auto-detect daily note settings (always attempt this)
	fmt.Printf("🔍 Scanning vault for daily notes...\n")
	
	// Try to auto-detect date format from existing files
	vault := obsidian.NewVault(vaultPath, dailyNotesDir, dateFormat)
	detectedFormat, err := vault.DetectDateFormat()
	
	if err == nil {
		fmt.Printf("📅 Found existing format: %s (%s directory)\n", detectedFormat, dailyNotesDir)
		
		if interactive {
			fmt.Print("Use detected format? (Y/n): ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(response)
			if response == "" || strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
				dateFormat = detectedFormat
			} else {
				// Still run interactive setup for manual configuration
				var err error
				dailyNotesDir, dateFormat, err = promptForDailyNoteConfig(vaultPath, dailyNotesDir, dateFormat)
				if err != nil {
					return fmt.Errorf("interactive setup failed: %w", err)
				}
			}
		} else {
			// Non-interactive: use detected format automatically
			dateFormat = detectedFormat
		}
	} else {
		fmt.Printf("⚠️  No existing daily notes found: %v\n", err)
		if interactive {
			fmt.Printf("📝 Setting up daily note configuration...\n")
			var err error
			dailyNotesDir, dateFormat, err = promptForDailyNoteConfig(vaultPath, dailyNotesDir, dateFormat)
			if err != nil {
				return fmt.Errorf("interactive setup failed: %w", err)
			}
		} else {
			fmt.Printf("📝 Using defaults: %s directory, %s format\n", dailyNotesDir, dateFormat)
		}
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
			"daily_notes_dir":  dailyNotesDir,
			"date_format":      dateFormat,
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
	fmt.Printf("📝 Daily notes directory: %s\n", dailyNotesDir)
	fmt.Printf("📅 Date format: %s\n", dateFormat)
	if len(projectDirs) > 0 {
		fmt.Printf("🚀 Project directories: %v\n", projectDirs)
	}
	fmt.Printf("\n⚡ Ready to use: obsidian-cli log\n")

	return nil
}

func promptForDailyNoteConfig(vaultPath, currentDailyNotesDir, currentDateFormat string) (string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("\n🔧 Interactive Daily Note Configuration")
	fmt.Println("This will help configure obsidian-cli to work with your specific daily note setup.")
	
	// Scan vault for existing daily notes to suggest configuration
	suggestions := scanVaultForDailyNotes(vaultPath)
	if len(suggestions) > 0 {
		fmt.Printf("\n📊 Found %d existing daily notes in your vault. Here are some patterns:\n", len(suggestions))
		for i, suggestion := range suggestions[:min(5, len(suggestions))] {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
	}
	
	// Ask for daily notes directory
	fmt.Printf("\n📁 Daily notes directory (current: %s): ", currentDailyNotesDir)
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
	fmt.Println("\n📅 Common daily note date formats:")
	fmt.Println("  1. YYYY-MM-DD-dddd          → 2025-07-19-Saturday")
	fmt.Println("  2. YYYY-MM-DD               → 2025-07-19")
	fmt.Println("  3. DD-MM-YYYY               → 19-07-2025")
	fmt.Println("  4. YYYY/MM/DD               → 2025/07/19")
	fmt.Println("  5. MMMM DD, YYYY            → July 19, 2025")
	fmt.Println("  6. DD MMMM YYYY             → 19 July 2025")
	fmt.Println("  7. YYYY-MM-DD dddd          → 2025-07-19 Saturday")
	fmt.Println("  8. YY-MM-DD                 → 25-07-19")
	
	fmt.Printf("\nCurrent format: %s\n", currentDateFormat)
	fmt.Print("Enter your date format (press Enter to keep current, or type a number 1-8 for common formats): ")
	
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
		dateFormat = "YYYY/MM/DD"
	case "5":
		dateFormat = "MMMM DD, YYYY"
	case "6":
		dateFormat = "DD MMMM YYYY"
	case "7":
		dateFormat = "YYYY-MM-DD dddd"
	case "8":
		dateFormat = "YY-MM-DD"
	default:
		dateFormat = dateFormatInput
	}
	
	// Validate the configuration
	fmt.Printf("\n✅ Configuration Summary:\n")
	fmt.Printf("   📁 Daily notes directory: %s\n", dailyNotesDir)
	fmt.Printf("   📅 Date format: %s\n", dateFormat)
	
	// Show what today's note would be named
	exampleName := formatDateExample(dateFormat)
	fmt.Printf("   📝 Today's note would be: %s/%s.md\n", dailyNotesDir, exampleName)
	
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
		return "2025-07-19-Saturday"
	case "YYYY-MM-DD":
		return "2025-07-19"
	case "DD-MM-YYYY":
		return "19-07-2025"
	case "YYYY/MM/DD":
		return "2025/07/19"
	case "MMMM DD, YYYY":
		return "July 19, 2025"
	case "DD MMMM YYYY":
		return "19 July 2025"
	case "YYYY-MM-DD dddd":
		return "2025-07-19 Saturday"
	case "YY-MM-DD":
		return "25-07-19"
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
