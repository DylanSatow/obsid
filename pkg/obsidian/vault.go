package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Vault struct {
	Path          string
	DailyNotesDir string
	DateFormat    string
}

func NewVault(path, dailyNotesDir, dateFormat string) *Vault {
	return &Vault{
		Path:          path,
		DailyNotesDir: dailyNotesDir,
		DateFormat:    dateFormat,
	}
}

func (v *Vault) GetDailyNotePath(date time.Time) string {
	// Convert date format to Go time format
	goFormat := convertDateFormatToGo(v.DateFormat)
	filename := date.Format(goFormat) + ".md"
	return filepath.Join(v.Path, v.DailyNotesDir, filename)
}

// convertDateFormatToGo converts common date formats to Go time format
func convertDateFormatToGo(format string) string {
	switch format {
	case "YYYY-MM-DD-dddd":
		return "2006-01-02-Monday"
	case "YYYY-MM-DD":
		return "2006-01-02"
	case "DD-MM-YYYY":
		return "02-01-2006"
	case "MM-DD-YYYY":
		return "01-02-2006"
	case "MM-DD-YY":
		return "01-02-06"
	case "YYYY/MM/DD":
		return "2006/01/02"
	case "MMMM DD, YYYY":
		return "January 02, 2006"
	case "DD MMMM YYYY":
		return "02 January 2006"
	case "YYYY-MM-DD dddd":
		return "2006-01-02 Monday"
	case "YY-MM-DD":
		return "06-01-02"
	default:
		// If we don't recognize the format, try to convert it
		// This is a basic conversion - could be enhanced
		goFormat := format
		goFormat = strings.ReplaceAll(goFormat, "YYYY", "2006")
		goFormat = strings.ReplaceAll(goFormat, "MM", "01")
		goFormat = strings.ReplaceAll(goFormat, "DD", "02")
		goFormat = strings.ReplaceAll(goFormat, "dddd", "Monday")
		goFormat = strings.ReplaceAll(goFormat, "MMMM", "January")
		goFormat = strings.ReplaceAll(goFormat, "YY", "06")
		return goFormat
	}
}

func (v *Vault) DailyNoteExists(date time.Time) bool {
	notePath := v.GetDailyNotePath(date)
	_, err := os.Stat(notePath)
	return err == nil
}

func (v *Vault) CreateDailyNote(date time.Time) error {
	notePath := v.GetDailyNotePath(date)

	// Create directory if it doesn't exist
	dir := filepath.Dir(notePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create file
	content := fmt.Sprintf("# %s\n\n", date.Format("Monday, January 2, 2006"))
	return os.WriteFile(notePath, []byte(content), 0644)
}

func (v *Vault) EnsureDailyNote(date time.Time) error {
	if !v.DailyNoteExists(date) {
		return v.CreateDailyNote(date)
	}
	return nil
}

func (v *Vault) Exists() bool {
	_, err := os.Stat(v.Path)
	return err == nil
}

// DetectDateFormat scans the daily notes directory for existing files and detects the date format
func (v *Vault) DetectDateFormat() (string, error) {
	dailyNotesPath := filepath.Join(v.Path, v.DailyNotesDir)
	
	// Check if daily notes directory exists
	if _, err := os.Stat(dailyNotesPath); os.IsNotExist(err) {
		return "", fmt.Errorf("daily notes directory not found: %s", dailyNotesPath)
	}

	// Read directory contents
	files, err := os.ReadDir(dailyNotesPath)
	if err != nil {
		return "", err
	}

	// Collect .md files that look like dates
	var dateFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			dateFiles = append(dateFiles, strings.TrimSuffix(file.Name(), ".md"))
		}
	}

	if len(dateFiles) == 0 {
		return "", fmt.Errorf("no daily note files found in %s", dailyNotesPath)
	}

	// Try to detect format from existing files
	detectedFormat := detectDateFormatFromFiles(dateFiles)
	return detectedFormat, nil
}

// detectDateFormatFromFiles analyzes filenames to determine date format
func detectDateFormatFromFiles(filenames []string) string {
	// Define clear, unambiguous patterns first
	patterns := map[string]string{
		`^\d{4}-\d{2}-\d{2}-\w+$`:     "YYYY-MM-DD-dddd",    // 2025-07-19-Saturday
		`^\d{4}-\d{2}-\d{2} \w+$`:     "YYYY-MM-DD dddd",    // 2025-07-19 Saturday
		`^\d{4}-\d{2}-\d{2}$`:         "YYYY-MM-DD",         // 2025-07-19
		`^\d{4}/\d{2}/\d{2}$`:         "YYYY/MM/DD",         // 2025/07/19
		`^[A-Z][a-z]+ \d{1,2}, \d{4}$`: "MMMM DD, YYYY",    // July 19, 2025
		`^\d{1,2} [A-Z][a-z]+ \d{4}$`: "DD MMMM YYYY",      // 19 July 2025
		`^\d{2}-\d{2}-\d{2}$`:         "YY-MM-DD",           // 25-07-19
		`^\d{1,2}-\d{1,2}-\d{2}$`:     "MM-DD-YY",           // 7-20-25 or 07-20-25
	}

	// Count matches for each pattern
	patternCounts := make(map[string]int)
	ambiguousFiles := []string{}
	
	for _, filename := range filenames {
		matched := false
		for pattern, format := range patterns {
			if matched, _ := regexp.MatchString(pattern, filename); matched {
				patternCounts[format]++
				matched = true
				break
			}
		}
		
		// Handle ambiguous MM-DD-YYYY vs DD-MM-YYYY case
		if !matched {
			if matched, _ := regexp.MatchString(`^\d{1,2}-\d{1,2}-\d{4}$`, filename); matched {
				ambiguousFiles = append(ambiguousFiles, filename)
			}
		}
	}

	// Resolve ambiguous MM-DD-YYYY vs DD-MM-YYYY files
	if len(ambiguousFiles) > 0 {
		format := resolveAmbiguousDateFormat(ambiguousFiles)
		patternCounts[format] += len(ambiguousFiles)
	}

	// Find the most common format
	maxCount := 0
	bestFormat := "YYYY-MM-DD" // default fallback
	
	for format, count := range patternCounts {
		if count > maxCount {
			maxCount = count
			bestFormat = format
		}
	}

	return bestFormat
}

// resolveAmbiguousDateFormat resolves between MM-DD-YYYY and DD-MM-YYYY formats
// by analyzing the values to see which interpretation makes more sense
func resolveAmbiguousDateFormat(filenames []string) string {
	ddmmCount := 0
	mmddCount := 0
	
	for _, filename := range filenames {
		// Parse the numbers
		parts := strings.Split(filename, "-")
		if len(parts) != 3 {
			continue
		}
		
		first := parts[0]
		second := parts[1]
		
		// Convert to integers
		firstInt := 0
		secondInt := 0
		fmt.Sscanf(first, "%d", &firstInt)
		fmt.Sscanf(second, "%d", &secondInt)
		
		// Check if first value could be a day (> 12) - indicates DD-MM-YYYY
		if firstInt > 12 {
			ddmmCount++
		}
		// Check if second value could be a day (> 12) - indicates MM-DD-YYYY  
		if secondInt > 12 {
			mmddCount++
		}
	}
	
	// If we found evidence for one format, use it
	if ddmmCount > mmddCount {
		return "DD-MM-YYYY"
	} else if mmddCount > ddmmCount {
		return "MM-DD-YYYY"
	}
	
	// If no clear evidence, default to MM-DD-YYYY (American format)
	return "MM-DD-YYYY"
}

// FindExistingDailyNote looks for an existing daily note for the given date using any detected format
func (v *Vault) FindExistingDailyNote(date time.Time) (string, bool) {
	dailyNotesPath := filepath.Join(v.Path, v.DailyNotesDir)
	
	// Try different common formats
	formats := []string{
		"YYYY-MM-DD",         // 2025-07-19
		"YYYY-MM-DD-dddd",    // 2025-07-19-Saturday
		"YYYY-MM-DD dddd",    // 2025-07-19 Saturday
		"DD-MM-YYYY",         // 19-07-2025
		"MM-DD-YYYY",         // 07-20-2025
		"MM-DD-YY",           // 07-20-25
		"YYYY/MM/DD",         // 2025/07/19
		"MMMM DD, YYYY",      // July 19, 2025
		"DD MMMM YYYY",       // 19 July 2025
		"YY-MM-DD",           // 25-07-19
	}
	
	for _, format := range formats {
		goFormat := convertDateFormatToGo(format)
		filename := date.Format(goFormat) + ".md"
		fullPath := filepath.Join(dailyNotesPath, filename)
		
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, true
		}
	}
	
	return "", false
}