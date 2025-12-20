package model

import "time"

// Schedule represents a schedule/task loaded from a child Notion database
type Schedule struct {
	ID              string
	Title           string
	DueDate         time.Time
	Description     string
	MessageTemplate string
	ReminderTimings []string
	NotionURL       string
	Properties      map[string]interface{} // All properties for template rendering
}

// Validate checks if the schedule is valid
func (s *Schedule) Validate() error {
	if s.Title == "" {
		return &ValidationError{Field: "Title", Message: "required"}
	}
	if s.DueDate.IsZero() {
		return &ValidationError{Field: "DueDate", Message: "required"}
	}
	return nil
}
