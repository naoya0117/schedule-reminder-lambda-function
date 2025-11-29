package service

import (
	"fmt"
	"schedule-reminder/internal/domain/calculator"
	"schedule-reminder/internal/domain/model"
	"strings"
)

// BuildMessage builds a notification message from template
func BuildMessage(schedule *model.Schedule, config *model.ReminderConfig, timing string) string {
	template := config.MessageTemplate
	if template == "" {
		// Default template
		template = "【リマインド】{title}\n期限: {due_date} ({days_text})\n{url}"
	}

	message := template

	// Replace variables
	message = strings.ReplaceAll(message, "{title}", schedule.Title)
	message = strings.ReplaceAll(message, "{due_date}", schedule.DueDate.Format("2006-01-02"))
	message = strings.ReplaceAll(message, "{days_text}", calculator.FormatDaysText(timing))
	message = strings.ReplaceAll(message, "{url}", schedule.NotionURL)
	message = strings.ReplaceAll(message, "{description}", schedule.Description)

	// Replace custom properties
	for key, value := range schedule.Properties {
		placeholder := fmt.Sprintf("{%s}", strings.ToLower(key))
		if value != nil {
			message = strings.ReplaceAll(message, placeholder, fmt.Sprintf("%v", value))
		}
	}

	return message
}
