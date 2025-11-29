package calculator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseAndCalculateReminderDate parses a timing string and calculates the reminder date
// Supported formats:
// - "当日" -> same day as due date
// - "1日前", "7日前" -> N days before
// - "1営業日前", "4営業日前" -> N business days before
// - "1週間前", "2週間前" -> N weeks before
func ParseAndCalculateReminderDate(dueDate time.Time, timing string, calculator *BusinessDayCalculator) (time.Time, error) {
	timing = strings.TrimSpace(timing)

	switch timing {
	case "当日":
		return dueDate, nil
	}

	// Try to match "N日前"
	if match := regexp.MustCompile(`^(\d+)日前$`).FindStringSubmatch(timing); len(match) == 2 {
		days, err := strconv.Atoi(match[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid number in timing: %s", timing)
		}
		return dueDate.AddDate(0, 0, -days), nil
	}

	// Try to match "N営業日前"
	if match := regexp.MustCompile(`^(\d+)営業日前$`).FindStringSubmatch(timing); len(match) == 2 {
		days, err := strconv.Atoi(match[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid number in timing: %s", timing)
		}
		if calculator == nil {
			return time.Time{}, fmt.Errorf("business day calculator required for: %s", timing)
		}
		return calculator.SubtractBusinessDays(dueDate, days), nil
	}

	// Try to match "N週間前"
	if match := regexp.MustCompile(`^(\d+)週間前$`).FindStringSubmatch(timing); len(match) == 2 {
		weeks, err := strconv.Atoi(match[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid number in timing: %s", timing)
		}
		return dueDate.AddDate(0, 0, -weeks*7), nil
	}

	return time.Time{}, fmt.Errorf("unsupported timing format: %s", timing)
}

// IsSameDate checks if two dates are on the same day (ignoring time)
func IsSameDate(d1, d2 time.Time) bool {
	y1, m1, day1 := d1.Date()
	y2, m2, day2 := d2.Date()
	return y1 == y2 && m1 == m2 && day1 == day2
}

// FormatDaysText formats the number of days into human-readable text
func FormatDaysText(timing string) string {
	switch timing {
	case "当日":
		return "今日"
	case "1日前":
		return "明日"
	case "2日前":
		return "2日後"
	default:
		// Extract number for other cases
		if match := regexp.MustCompile(`^(\d+)`).FindStringSubmatch(timing); len(match) == 2 {
			return match[1] + "日後"
		}
		return timing
	}
}
