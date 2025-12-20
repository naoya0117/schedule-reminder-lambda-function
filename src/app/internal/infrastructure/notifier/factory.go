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

	// TODO: Implement additional Slack features (attachments, blocks)
	case "line":
		if config.ChannelToken == "" {
			return nil, fmt.Errorf("channel access token required for LINE")
		}
		if config.LineRecipientID == "" {
			return nil, fmt.Errorf("line recipient ID required for LINE")
		}
		return NewLineNotifier(config.ChannelToken), nil
	case "slack":
		if config.WebhookURL == "" {
			return nil, fmt.Errorf("webhook URL required for Slack")
		}
		return NewSlackNotifier(config.WebhookURL), nil

	default:
		return nil, fmt.Errorf("unsupported notification channel: %s", config.NotificationChannel)
	}
}
