package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
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




