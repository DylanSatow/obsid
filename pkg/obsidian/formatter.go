package obsidian

import (
	"fmt"
	"strings"

	"github.com/DylanSatow/obsidian-cli/pkg/git"
)

type ProjectEntry struct {
	ProjectName string
	TimeRange   string
	Branch      string
	Commits     []git.Commit
	Files       []string
	Tags        []string
}

func FormatProjectEntry(repo *git.Repository, commits []git.Commit, files []string, timeRange string) string {
	var sb strings.Builder

	// Header with metadata
	sb.WriteString(fmt.Sprintf("**Time**: %s\n", timeRange))
	sb.WriteString(fmt.Sprintf("**Branch**: %s\n\n", repo.Branch))

	// Commits section
	if len(commits) > 0 {
		sb.WriteString("#### Recent Commits\n")
		for _, commit := range commits {
			timestamp := commit.Timestamp.Format("15:04")
			sb.WriteString(fmt.Sprintf("- `%s` (%s)\n", commit.Message, timestamp))
		}
		sb.WriteString("\n")
	}

	// Changes section
	if len(files) > 0 {
		sb.WriteString("#### Key Changes\n")
		sb.WriteString(fmt.Sprintf("**Files Modified**: %d files\n\n", len(files)))

		if len(files) <= 5 {
			for _, file := range files {
				sb.WriteString(fmt.Sprintf("- %s\n", file))
			}
		} else {
			for _, file := range files[:3] {
				sb.WriteString(fmt.Sprintf("- %s\n", file))
			}
			sb.WriteString(fmt.Sprintf("- ... and %d more files\n", len(files)-3))
		}
		sb.WriteString("\n")
	}

	// Tags
	sb.WriteString("**Tags**: #programming #")
	sb.WriteString(strings.ReplaceAll(repo.Name, "-", "_"))
	sb.WriteString("\n\n---\n\n")

	return sb.String()
}