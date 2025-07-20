package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
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
	// Convert from YYYY-MM-DD-dddd format to Go time format
	goFormat := "2006-01-02-Monday"
	filename := date.Format(goFormat) + ".md"
	return filepath.Join(v.Path, v.DailyNotesDir, filename)
}

func (v *Vault) EnsureDailyNote(date time.Time) error {
	notePath := v.GetDailyNotePath(date)

	// Create directory if it doesn't exist
	dir := filepath.Dir(notePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create file if it doesn't exist
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		content := fmt.Sprintf("# %s\n\n", date.Format("Monday, January 2, 2006"))
		return os.WriteFile(notePath, []byte(content), 0644)
	}

	return nil
}

func (v *Vault) Exists() bool {
	_, err := os.Stat(v.Path)
	return err == nil
}