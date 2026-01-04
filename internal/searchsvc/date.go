package searchsvc

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// ParseDate parses relative and absolute date strings.
//
// This intentionally preserves the parsing behavior that previously lived in pkg/commands/search.go.
func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	dateStr = strings.TrimSpace(dateStr)
	dateStrLower := strings.ToLower(dateStr)

	now := time.Now()

	switch dateStrLower {
	case "today":
		return startOfDay(now), nil
	case "yesterday":
		return startOfDay(now.AddDate(0, 0, -1)), nil
	case "last week", "lastweek":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		thisWeekStart := now.AddDate(0, 0, -(weekday - 1))
		return startOfDay(thisWeekStart.AddDate(0, 0, -7)), nil
	case "this week", "thisweek":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return startOfDay(now.AddDate(0, 0, -(weekday - 1))), nil
	case "this month", "thismonth":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), nil
	case "last month", "lastmonth":
		lastMonth := now.AddDate(0, -1, 0)
		return time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location()), nil
	case "last year", "lastyear":
		return time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location()), nil
	case "this year", "thisyear":
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()), nil
	}

	// Relative date patterns
	re := regexp.MustCompile(`^(\d+)\s+(day|days|week|weeks|month|months|year|years)\s+ago$`)
	if matches := re.FindStringSubmatch(dateStrLower); len(matches) == 3 {
		n, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, err
		}
		unit := matches[2]
		switch unit {
		case "day", "days":
			return startOfDay(now.AddDate(0, 0, -n)), nil
		case "week", "weeks":
			return startOfDay(now.AddDate(0, 0, -7*n)), nil
		case "month", "months":
			return startOfDay(now.AddDate(0, -n, 0)), nil
		case "year", "years":
			return startOfDay(now.AddDate(-n, 0, 0)), nil
		}
	}

	// Handle "last <unit>" without numbers (kept for compatibility with the previous implementation).
	if strings.HasPrefix(dateStrLower, "last ") {
		rest := strings.TrimPrefix(dateStrLower, "last ")
		switch rest {
		case "week":
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			thisWeekStart := now.AddDate(0, 0, -(weekday - 1))
			return startOfDay(thisWeekStart.AddDate(0, 0, -7)), nil
		case "month":
			thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			return thisMonthStart.AddDate(0, -1, 0), nil
		case "year":
			return time.Date(now.Year()-1, 1, 1, 0, 0, 0, 0, now.Location()), nil
		}
	}

	// Try common absolute formats.
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, dateStr, now.Location()); err == nil {
			return t, nil
		}
	}

	return time.Time{}, &time.ParseError{Layout: "relative date or common absolute format", Value: dateStr, LayoutElem: "", ValueElem: "", Message: "unrecognized date format"}
}
