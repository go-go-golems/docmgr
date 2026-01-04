package searchsvc

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

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
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "yesterday":
		y := now.AddDate(0, 0, -1)
		return time.Date(y.Year(), y.Month(), y.Day(), 0, 0, 0, 0, y.Location()), nil
	case "last week", "lastweek":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		lastWeekStart := now.AddDate(0, 0, -weekday-6)
		return time.Date(lastWeekStart.Year(), lastWeekStart.Month(), lastWeekStart.Day(), 0, 0, 0, 0, now.Location()), nil
	case "this month", "thismonth":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), nil
	case "last month", "lastmonth":
		lastMonth := now.AddDate(0, -1, 0)
		return time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location()), nil
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
			return now.AddDate(0, 0, -n), nil
		case "week", "weeks":
			return now.AddDate(0, 0, -7*n), nil
		case "month", "months":
			return now.AddDate(0, -n, 0), nil
		case "year", "years":
			return now.AddDate(-n, 0, 0), nil
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
