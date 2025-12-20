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

// SlackNotifier sends notifications via Slack Incoming Webhooks.
type SlackNotifier struct {
	webhookURL string
	httpClient *http.Client
}

// NewSlackNotifier creates a new Slack notifier.
func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send sends a notification to Slack.
func (s *SlackNotifier) Send(ctx context.Context, notification *model.Notification) error {
	payload := map[string]string{
		"text": notification.Message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// Type returns the notifier type.
func (s *SlackNotifier) Type() string {
	return "Slack"
}
