package notion

import (
	"context"
	"fmt"
	"schedule-reminder/internal/domain/model"
	"time"

	"github.com/jomei/notionapi"
)

// Client wraps the Notion API client
type Client struct {
	client *notionapi.Client
}

// NewClient creates a new Notion client
func NewClient(apiKey string) *Client {
	return &Client{
		client: notionapi.NewClient(notionapi.Token(apiKey)),
	}
}

// LoadReminderConfigs loads all reminder configurations from the master database
func (c *Client) LoadReminderConfigs(ctx context.Context, masterDBID string) ([]*model.ReminderConfig, error) {
	query := &notionapi.DatabaseQueryRequest{
		Filter: &notionapi.PropertyFilter{
			Property: "Enabled",
			Checkbox: &notionapi.CheckboxFilterCondition{
				Equals: true,
			},
		},
	}

	result, err := c.client.Database.Query(ctx, notionapi.DatabaseID(masterDBID), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query master database: %w", err)
	}

	var configs []*model.ReminderConfig
	for _, page := range result.Results {
		config, err := c.parseReminderConfig(page)
		if err != nil {
			// Log error but continue processing other configs
			fmt.Printf("Warning: failed to parse config %s: %v\n", page.ID, err)
			continue
		}

		if err := config.Validate(); err != nil {
			fmt.Printf("Warning: invalid config %s: %v\n", page.ID, err)
			continue
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// parseReminderConfig extracts configuration from a Notion page
func (c *Client) parseReminderConfig(page notionapi.Page) (*model.ReminderConfig, error) {
	config := &model.ReminderConfig{
		ID: page.ID.String(),
	}

	// Name (Title)
	if titleProp, ok := page.Properties["Name"].(*notionapi.TitleProperty); ok && len(titleProp.Title) > 0 {
		config.Name = titleProp.Title[0].PlainText
	}

	// Target Database ID
	if textProp, ok := page.Properties["Target Database ID"].(*notionapi.RichTextProperty); ok && len(textProp.RichText) > 0 {
		config.TargetDatabaseID = textProp.RichText[0].PlainText
	}

	// Reminder Timings (Multi-select)
	if multiSelectProp, ok := page.Properties["Reminder Timings"].(*notionapi.MultiSelectProperty); ok {
		for _, option := range multiSelectProp.MultiSelect {
			config.ReminderTimings = append(config.ReminderTimings, option.Name)
		}
	}

	// Notification Channel (Select)
	if selectProp, ok := page.Properties["Notification Channel"].(*notionapi.SelectProperty); ok && selectProp.Select != nil {
		config.NotificationChannel = selectProp.Select.Name
	}

	// Webhook URL
	if urlProp, ok := page.Properties["Webhook URL"].(*notionapi.URLProperty); ok {
		config.WebhookURL = string(urlProp.URL)
	}

	// Channel Token
	if textProp, ok := page.Properties["Channel Access Token"].(*notionapi.RichTextProperty); ok && len(textProp.RichText) > 0 {
		config.ChannelToken = textProp.RichText[0].PlainText
	}

	// Message Template
	if textProp, ok := page.Properties["Message Template"].(*notionapi.RichTextProperty); ok && len(textProp.RichText) > 0 {
		config.MessageTemplate = textProp.RichText[0].PlainText
	}

	// Date Property Name
	config.DatePropertyName = "Due Date" // Default
	if textProp, ok := page.Properties["Date Property Name"].(*notionapi.RichTextProperty); ok && len(textProp.RichText) > 0 {
		config.DatePropertyName = textProp.RichText[0].PlainText
	}

	// Title Property Name
	config.TitlePropertyName = "Title" // Default
	if textProp, ok := page.Properties["Title Property Name"].(*notionapi.RichTextProperty); ok && len(textProp.RichText) > 0 {
		config.TitlePropertyName = textProp.RichText[0].PlainText
	}

	// Timezone
	timezone := "Asia/Tokyo" // Default
	if selectProp, ok := page.Properties["Timezone"].(*notionapi.SelectProperty); ok && selectProp.Select != nil {
		timezone = selectProp.Select.Name
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.FixedZone("JST", 9*3600) // Fallback to JST
	}
	config.Timezone = loc

	return config, nil
}
