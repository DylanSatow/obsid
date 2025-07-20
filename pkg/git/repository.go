package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Repository struct {
	Path   string
	Name   string
	Branch string
}

type Commit struct {
	Hash      string
	Message   string
	Author    string
	Timestamp time.Time
	Files     []string
}

func FindRepository(startPath string) (*Repository, error) {
	dir, err := filepath.Abs(startPath)
	if err != nil {
		return nil, err
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			name := filepath.Base(dir)
			branch, _ := getCurrentBranch(dir)
			return &Repository{
				Path:   dir,
				Name:   name,
				Branch: branch,
			}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf("not a git repository")
}

func getCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (r *Repository) GetCommits(since time.Time, maxCommits int) ([]Commit, error) {
	sinceStr := since.Format("2006-01-02 15:04:05")
	cmd := exec.Command("git", "log",
		"--since="+sinceStr,
		"--pretty=format:%H|%s|%an|%ad",
		"--date=iso",
		fmt.Sprintf("--max-count=%d", maxCommits))
	cmd.Dir = r.Path

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []Commit
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "|")
		if len(parts) != 4 {
			continue
		}

		timestamp, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])
		commits = append(commits, Commit{
			Hash:      parts[0],
			Message:   parts[1],
			Author:    parts[2],
			Timestamp: timestamp,
		})
	}

	return commits, nil
}

func (r *Repository) GetChangedFiles(since time.Time) ([]string, error) {
	sinceStr := since.Format("2006-01-02 15:04:05")
	cmd := exec.Command("git", "diff", "--name-only", "--since="+sinceStr, "HEAD")
	cmd.Dir = r.Path

	output, err := cmd.Output()
	if err != nil {
		// If git diff --since fails, try a different approach
		cmd = exec.Command("git", "log", "--name-only", "--pretty=format:", "--since="+sinceStr)
		cmd.Dir = r.Path
		output, err = cmd.Output()
		if err != nil {
			return nil, err
		}
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			files = append(files, line)
		}
	}

	return removeDuplicates(files), nil
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}