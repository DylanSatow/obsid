package obsidian

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func (v *Vault) AppendProjectEntry(date time.Time, projectName string, content string) error {
	notePath := v.GetDailyNotePath(date)

	// Read existing content
	file, err := os.Open(notePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Find or create Projects section
	projectsIndex := findProjectsSection(lines)
	if projectsIndex == -1 {
		// Add Projects section
		lines = append(lines, "", "## Projects", "")
		projectsIndex = len(lines) - 1
	}

	// Find existing project entry or determine where to insert
	insertIndex := findProjectInsertionPoint(lines, projectsIndex, projectName)

	projectEntry := formatProjectEntry(projectName, content)
	newLines := insertLines(lines, insertIndex, strings.Split(projectEntry, "\n"))

	// Write back to file
	return os.WriteFile(notePath, []byte(strings.Join(newLines, "\n")), 0644)
}

func findProjectsSection(lines []string) int {
	for i, line := range lines {
		if strings.HasPrefix(line, "## Projects") {
			return i
		}
	}
	return -1
}

func findProjectInsertionPoint(lines []string, projectsIndex int, projectName string) int {
	// Look for existing project entry
	for i := projectsIndex + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "### "+projectName) {
			// Found existing entry - replace from here
			return i
		}
		if strings.HasPrefix(lines[i], "## ") {
			// Hit next section - insert before it
			return i
		}
	}
	// Insert at end
	return len(lines)
}

func formatProjectEntry(projectName, content string) string {
	return fmt.Sprintf("### %s\n%s", projectName, content)
}

func insertLines(lines []string, index int, newLines []string) []string {
	// If we're replacing an existing project entry, we need to find where it ends
	if index < len(lines) && strings.HasPrefix(lines[index], "### ") {
		// Find end of existing project entry
		endIndex := index + 1
		for endIndex < len(lines) {
			if strings.HasPrefix(lines[endIndex], "### ") || strings.HasPrefix(lines[endIndex], "## ") {
				break
			}
			endIndex++
		}
		// Replace the existing entry
		result := make([]string, 0, len(lines)-((endIndex-index))+len(newLines))
		result = append(result, lines[:index]...)
		result = append(result, newLines...)
		result = append(result, lines[endIndex:]...)
		return result
	}

	// Insert new entry
	result := make([]string, 0, len(lines)+len(newLines))
	result = append(result, lines[:index]...)
	result = append(result, newLines...)
	result = append(result, lines[index:]...)
	return result
}