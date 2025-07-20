package obsidian

import (
	"fmt"
	"strings"

	"github.com/DylanSatow/obsidian-cli/pkg/git"
)

func FormatProjectEntry(repo *git.Repository, commits []git.Commit, files []string, timeRange string) string {
	var sb strings.Builder

	// Clean, focused work log format
	sb.WriteString(fmt.Sprintf("**%s** • %s", timeRange, formatWorkSummary(commits, files)))
	sb.WriteString("\n\n")

	// What I accomplished (derived from commit messages)
	if len(commits) > 0 {
		accomplishments := extractAccomplishments(commits)
		if len(accomplishments) > 0 {
			for _, accomplishment := range accomplishments {
				sb.WriteString(fmt.Sprintf("- %s\n", accomplishment))
			}
			sb.WriteString("\n")
		}
	}

	// Key areas worked on (files grouped by functionality)
	if len(files) > 0 {
		areas := groupFilesByArea(files)
		if len(areas) > 0 {
			sb.WriteString("**Areas:** ")
			sb.WriteString(strings.Join(areas, ", "))
			sb.WriteString("\n\n")
		}
	}

	// Simple tag
	sb.WriteString(fmt.Sprintf("#%s\n\n", cleanProjectName(repo.Name)))

	return sb.String()
}

// formatWorkSummary creates a concise summary of the work session
func formatWorkSummary(commits []git.Commit, files []string) string {
	if len(commits) == 0 && len(files) == 0 {
		return "code review/exploration"
	}

	var parts []string

	if len(commits) > 0 {
		if len(commits) == 1 {
			parts = append(parts, "1 commit")
		} else {
			parts = append(parts, fmt.Sprintf("%d commits", len(commits)))
		}
	}

	if len(files) > 0 {
		if len(files) <= 3 {
			parts = append(parts, fmt.Sprintf("%d files", len(files)))
		} else {
			parts = append(parts, fmt.Sprintf("%d+ files", len(files)))
		}
	}

	return strings.Join(parts, ", ")
}

// extractAccomplishments converts commit messages into meaningful accomplishments
func extractAccomplishments(commits []git.Commit) []string {
	var accomplishments []string
	
	for _, commit := range commits {
		accomplishment := cleanCommitMessage(commit.Message)
		if accomplishment != "" && !isDuplicateAccomplishment(accomplishment, accomplishments) {
			accomplishments = append(accomplishments, accomplishment)
		}
	}

	// Limit to most important accomplishments
	if len(accomplishments) > 4 {
		accomplishments = accomplishments[:4]
	}

	return accomplishments
}

// cleanCommitMessage converts technical commit messages to readable accomplishments
func cleanCommitMessage(message string) string {
	// Remove common prefixes
	message = strings.TrimSpace(message)
	
	// Convert technical prefixes to readable format
	prefixMap := map[string]string{
		"feat:":     "Added",
		"feature:":  "Added", 
		"fix:":      "Fixed",
		"bugfix:":   "Fixed",
		"refactor:": "Refactored",
		"docs:":     "Updated docs for",
		"test:":     "Added tests for",
		"style:":    "Improved styling of",
		"chore:":    "Updated",
		"update:":   "Updated",
		"add:":      "Added",
		"remove:":   "Removed",
		"delete:":   "Removed",
	}

	lower := strings.ToLower(message)
	for prefix, replacement := range prefixMap {
		if strings.HasPrefix(lower, prefix) {
			rest := strings.TrimSpace(message[len(prefix):])
			if rest != "" {
				// Clean up double words and ensure proper capitalization
				if len(rest) > 0 {
					rest = strings.ToLower(string(rest[0])) + rest[1:]
					// Remove redundant words like "add" in "Added add feature"
					rest = removeRedundantWords(rest, strings.ToLower(replacement))
				}
				return fmt.Sprintf("%s %s", replacement, rest)
			}
		}
	}

	// If no prefix found, capitalize first letter and return
	if len(message) > 0 {
		return strings.ToUpper(string(message[0])) + message[1:]
	}

	return message
}

// isDuplicateAccomplishment checks if an accomplishment is essentially the same as existing ones
func isDuplicateAccomplishment(new string, existing []string) bool {
	newClean := strings.ToLower(strings.TrimSpace(new))
	
	for _, exist := range existing {
		existClean := strings.ToLower(strings.TrimSpace(exist))
		
		// Check for substantial overlap in words
		if similarity := calculateSimilarity(newClean, existClean); similarity > 0.7 {
			return true
		}
	}
	
	return false
}

// calculateSimilarity returns a rough similarity score between two strings
func calculateSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	
	wordsA := strings.Fields(a)
	wordsB := strings.Fields(b)
	
	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0.0
	}
	
	matches := 0
	for _, wordA := range wordsA {
		for _, wordB := range wordsB {
			if wordA == wordB && len(wordA) > 2 { // Only count meaningful words
				matches++
				break
			}
		}
	}
	
	maxLen := len(wordsA)
	if len(wordsB) > maxLen {
		maxLen = len(wordsB)
	}
	
	return float64(matches) / float64(maxLen)
}

// groupFilesByArea organizes files into logical areas
func groupFilesByArea(files []string) []string {
	areas := make(map[string]bool)
	
	for _, file := range files {
		area := categorizeFile(file)
		if area != "" {
			areas[area] = true
		}
	}
	
	var result []string
	for area := range areas {
		result = append(result, area)
	}
	
	// Limit to avoid clutter
	if len(result) > 4 {
		result = result[:3]
		result = append(result, "...")
	}
	
	return result
}

// categorizeFile determines the functional area of a file
func categorizeFile(file string) string {
	file = strings.ToLower(file)
	
	// Frontend/UI
	if strings.Contains(file, "component") || strings.Contains(file, ".tsx") || 
	   strings.Contains(file, ".jsx") || strings.Contains(file, ".vue") ||
	   strings.Contains(file, "ui/") || strings.Contains(file, "frontend/") {
		return "frontend"
	}
	
	// Styling
	if strings.Contains(file, ".css") || strings.Contains(file, ".scss") || 
	   strings.Contains(file, ".sass") || strings.Contains(file, "style") {
		return "styling"
	}
	
	// Backend/API
	if strings.Contains(file, "api/") || strings.Contains(file, "server/") ||
	   strings.Contains(file, "backend/") || strings.Contains(file, "route") ||
	   strings.Contains(file, "controller") || strings.Contains(file, "handler") {
		return "backend"
	}
	
	// Database
	if strings.Contains(file, "database") || strings.Contains(file, "db/") ||
	   strings.Contains(file, "migration") || strings.Contains(file, "schema") ||
	   strings.Contains(file, ".sql") {
		return "database"
	}
	
	// Configuration
	if strings.Contains(file, "config") || strings.Contains(file, ".env") ||
	   strings.Contains(file, ".yml") || strings.Contains(file, ".yaml") ||
	   strings.Contains(file, ".json") && (strings.Contains(file, "package") || strings.Contains(file, "config")) {
		return "config"
	}
	
	// Tests
	if strings.Contains(file, "test") || strings.Contains(file, "spec") ||
	   strings.Contains(file, "__test__") {
		return "tests"
	}
	
	// Documentation
	if strings.Contains(file, "readme") || strings.Contains(file, ".md") ||
	   strings.Contains(file, "doc") {
		return "docs"
	}
	
	// Core logic (fallback for main implementation files)
	if strings.Contains(file, "main") || strings.Contains(file, "index") ||
	   strings.Contains(file, "app") || strings.Contains(file, "core") {
		return "core"
	}
	
	// If we can't categorize, use the directory name or file type
	parts := strings.Split(file, "/")
	if len(parts) > 1 {
		return parts[0] // Use top-level directory
	}
	
	// Use file extension as last resort
	if ext := getFileExtension(file); ext != "" {
		return ext
	}
	
	return ""
}

// getFileExtension returns a clean file extension
func getFileExtension(file string) string {
	parts := strings.Split(file, ".")
	if len(parts) > 1 {
		ext := parts[len(parts)-1]
		switch ext {
		case "js", "ts", "py", "go", "rb", "php", "java", "cpp", "c", "rs":
			return ext
		}
	}
	return ""
}

// cleanProjectName creates a clean tag from project name
func cleanProjectName(name string) string {
	// Convert to lowercase and replace special characters
	clean := strings.ToLower(name)
	clean = strings.ReplaceAll(clean, "-", "_")
	clean = strings.ReplaceAll(clean, " ", "_")
	clean = strings.ReplaceAll(clean, ".", "_")
	
	// Remove any non-alphanumeric characters except underscores
	var result strings.Builder
	for _, char := range clean {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			result.WriteRune(char)
		}
	}
	
	return result.String()
}

// removeRedundantWords removes redundant words from the beginning of a message
func removeRedundantWords(message, action string) string {
	words := strings.Fields(message)
	if len(words) == 0 {
		return message
	}
	
	// Check if first word is redundant with the action
	firstWord := strings.ToLower(words[0])
	actionLower := strings.ToLower(action)
	
	// Remove redundant words like "add" in "Added add feature"
	redundantWords := []string{"add", "added", "fix", "fixed", "update", "updated", "remove", "removed"}
	
	for _, redundant := range redundantWords {
		if firstWord == redundant && strings.Contains(actionLower, redundant) {
			// Remove the redundant word
			if len(words) > 1 {
				return strings.Join(words[1:], " ")
			}
			return ""
		}
	}
	
	return message
}