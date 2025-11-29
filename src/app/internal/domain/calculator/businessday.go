package calculator

import "time"

// BusinessDayCalculator calculates business days excluding weekends and holidays
type BusinessDayCalculator struct {
	holidays    map[string]bool // "2025-01-01" format for fast lookup
	weekendDays map[time.Weekday]bool
	timezone    *time.Location
}

// NewBusinessDayCalculator creates a new business day calculator
func NewBusinessDayCalculator(holidays []time.Time, timezone *time.Location) *BusinessDayCalculator {
	calc := &BusinessDayCalculator{
		holidays:    make(map[string]bool),
		weekendDays: map[time.Weekday]bool{time.Saturday: true, time.Sunday: true},
		timezone:    timezone,
	}

	// Convert holidays to map for O(1) lookup
	for _, holiday := range holidays {
		key := holiday.In(timezone).Format("2006-01-02")
		calc.holidays[key] = true
	}

	return calc
}

// IsBusinessDay checks if the given date is a business day
func (c *BusinessDayCalculator) IsBusinessDay(date time.Time) bool {
	date = date.In(c.timezone)

	// Check if weekend
	if c.weekendDays[date.Weekday()] {
		return false
	}

	// Check if holiday
	key := date.Format("2006-01-02")
	if c.holidays[key] {
		return false
	}

	return true
}

// SubtractBusinessDays subtracts N business days from the given date
func (c *BusinessDayCalculator) SubtractBusinessDays(date time.Time, days int) time.Time {
	current := date.In(c.timezone)
	remaining := days

	for remaining > 0 {
		current = current.AddDate(0, 0, -1)
		if c.IsBusinessDay(current) {
			remaining--
		}
	}

	return current
}

// AddBusinessDays adds N business days to the given date
func (c *BusinessDayCalculator) AddBusinessDays(date time.Time, days int) time.Time {
	current := date.In(c.timezone)
	remaining := days

	for remaining > 0 {
		current = current.AddDate(0, 0, 1)
		if c.IsBusinessDay(current) {
			remaining--
		}
	}

	return current
}
