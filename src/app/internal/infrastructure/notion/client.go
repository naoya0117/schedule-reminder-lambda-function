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

	var configs []*model.ReminderConfig
	for {
		result, err := c.client.Database.Query(ctx, notionapi.DatabaseID(masterDBID), query)
		if err != nil {
			return nil, fmt.Errorf("failed to query master database: %w", err)
		}

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

		if !result.HasMore || result.NextCursor == "" {
			break
		}
		query.StartCursor = result.NextCursor
	}

	return configs, nil
}

// parseReminderConfig extracts configuration from a Notion page
func (c *Client) parseReminderConfig(page notionapi.Page) (*model.ReminderConfig, error) {
	config := &model.ReminderConfig{
		ID: page.ID.String(),
	}

	// Name (Title)
	if titleProp := getTitleProperty(page, "名前", "Name"); titleProp != nil && len(titleProp.Title) > 0 {
		config.Name = titleProp.Title[0].PlainText
	}

	// Target Database ID
	if textProp := getRichTextProperty(page, "対象データベースID", "Target Database ID"); textProp != nil && len(textProp.RichText) > 0 {
		config.TargetDatabaseID = textProp.RichText[0].PlainText
	}

	// Reminder Timings (Multi-select)
	if multiSelectProp := getMultiSelectProperty(page, "リマインドタイミング", "Reminder Timings"); multiSelectProp != nil {
		for _, option := range multiSelectProp.MultiSelect {
			config.ReminderTimings = append(config.ReminderTimings, option.Name)
		}
	}

	// Notification Channel (Select)
	if selectProp := getSelectProperty(page, "通知チャネル", "Notification Channel"); selectProp != nil && selectProp.Select.Name != "" {
		config.NotificationChannel = selectProp.Select.Name
	}

	// Webhook URL
	if urlProp := getURLProperty(page, "Webhook URL"); urlProp != nil {
		config.WebhookURL = string(urlProp.URL)
	}

	// Channel Token
	if textProp := getRichTextProperty(page, "チャネルアクセストークン", "Channel Access Token"); textProp != nil && len(textProp.RichText) > 0 {
		config.ChannelToken = textProp.RichText[0].PlainText
	}

	// LINE Recipient ID
	if textProp := getRichTextProperty(page, "LINE送信先ID", "Line Recipient ID", "LINE Recipient ID"); textProp != nil && len(textProp.RichText) > 0 {
		config.LineRecipientID = textProp.RichText[0].PlainText
	}

	// Message Template
	if textProp := getRichTextProperty(page, "メッセージテンプレート", "Message Template"); textProp != nil && len(textProp.RichText) > 0 {
		config.MessageTemplate = textProp.RichText[0].PlainText
	}

	// Date Property Name
	config.DatePropertyName = "期限日" // Fixed

	// Title Property Name
	config.TitlePropertyName = "タイトル" // Fixed

	// Timezone
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		loc = time.FixedZone("JST", 9*3600) // Fallback to JST
	}
	config.Timezone = loc

	return config, nil
}

func getTitleProperty(page notionapi.Page, names ...string) *notionapi.TitleProperty {
	for _, name := range names {
		if prop, ok := page.Properties[name].(*notionapi.TitleProperty); ok {
			return prop
		}
	}
	return nil
}

func getRichTextProperty(page notionapi.Page, names ...string) *notionapi.RichTextProperty {
	for _, name := range names {
		if prop, ok := page.Properties[name].(*notionapi.RichTextProperty); ok {
			return prop
		}
	}
	return nil
}

func getMultiSelectProperty(page notionapi.Page, names ...string) *notionapi.MultiSelectProperty {
	for _, name := range names {
		if prop, ok := page.Properties[name].(*notionapi.MultiSelectProperty); ok {
			return prop
		}
	}
	return nil
}

func getSelectProperty(page notionapi.Page, names ...string) *notionapi.SelectProperty {
	for _, name := range names {
		if prop, ok := page.Properties[name].(*notionapi.SelectProperty); ok {
			return prop
		}
	}
	return nil
}

func getURLProperty(page notionapi.Page, names ...string) *notionapi.URLProperty {
	for _, name := range names {
		if prop, ok := page.Properties[name].(*notionapi.URLProperty); ok {
			return prop
		}
	}
	return nil
}
