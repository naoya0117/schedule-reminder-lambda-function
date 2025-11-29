package notifier

import (
	"fmt"
	"schedule-reminder/internal/domain/model"
	"strings"
)

// CreateNotifier creates a notifier based on the configuration
func CreateNotifier(config *model.ReminderConfig) (Notifier, error) {
	channel := strings.ToLower(config.NotificationChannel)

	switch channel {
	case "discord":
		if config.WebhookURL == "" {
			return nil, fmt.Errorf("webhook URL required for Discord")
		}
		return NewDiscordNotifier(config.WebhookURL), nil

	// TODO: Implement LINE and Slack notifiers
	case "line":
		return nil, fmt.Errorf("LINE notifier not yet implemented")
	case "slack":
		return nil, fmt.Errorf("Slack notifier not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported notification channel: %s", config.NotificationChannel)
	}
}
