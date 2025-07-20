package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseTimeframe parses timeframe strings like "1h", "2h", "today", "30m"
func ParseTimeframe(timeframe string) (time.Time, error) {
	switch strings.ToLower(timeframe) {
	case "today":
		return time.Now().Truncate(24 * time.Hour), nil
	case "yesterday":
		return time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour), nil
	}

	// Parse duration format like "1h", "30m", "2h30m"
	if strings.HasSuffix(timeframe, "h") || strings.HasSuffix(timeframe, "m") {
		duration, err := time.ParseDuration(timeframe)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid timeframe format: %s", timeframe)
		}
		return time.Now().Add(-duration), nil
	}

	// Try to parse as number of hours
	if hours, err := strconv.Atoi(timeframe); err == nil {
		return time.Now().Add(-time.Duration(hours) * time.Hour), nil
	}

	return time.Time{}, fmt.Errorf("unsupported timeframe format: %s", timeframe)
}

// FormatTimeRange creates a human-readable time range string
func FormatTimeRange(since time.Time) string {
	now := time.Now()
	duration := now.Sub(since)

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%s - %s (%dm)", since.Format("3:04PM"), now.Format("3:04PM"), minutes)
	}

	if duration < 24*time.Hour && since.Day() == now.Day() {
		return fmt.Sprintf("%s - %s", since.Format("3:04PM"), now.Format("3:04PM"))
	}

	if since.Day() == now.Day() {
		return fmt.Sprintf("Today %s - %s", since.Format("3:04PM"), now.Format("3:04PM"))
	}

	return fmt.Sprintf("%s - %s", since.Format("Jan 2 3:04PM"), now.Format("Jan 2 3:04PM"))
}