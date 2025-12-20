package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"schedule-reminder/internal/domain/model"
	"time"
)

const linePushEndpoint = "https://api.line.me/v2/bot/message/push"

// LineNotifier sends notifications via LINE Messaging API (push message).
type LineNotifier struct {
	channelToken string
	httpClient   *http.Client
}

// NewLineNotifier creates a new LINE notifier.
func NewLineNotifier(channelToken string) *LineNotifier {
	return &LineNotifier{
		channelToken: channelToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send sends a notification to LINE.
func (l *LineNotifier) Send(ctx context.Context, notification *model.Notification) error {
	if notification.Destination == "" {
		return fmt.Errorf("line recipient ID is required")
	}

	payload := map[string]interface{}{
		"to": notification.Destination,
		"messages": []map[string]string{
			{
				"type": "text",
				"text": notification.Message,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", linePushEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.channelToken)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("line push returned status %d", resp.StatusCode)
	}

	return nil
}

// Type returns the notifier type.
func (l *LineNotifier) Type() string {
	return "LINE"
}
