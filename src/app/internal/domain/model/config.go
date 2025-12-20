package model

import "time"

// ReminderConfig represents configuration loaded from the parent Notion database
type ReminderConfig struct {
	ID                  string
	Name                string
	TargetDatabaseID    string
	ReminderTimings     []string
	NotificationChannel string
	WebhookURL          string
	ChannelToken        string
	MessageTemplate     string
	DatePropertyName    string
	TitlePropertyName   string
	Timezone            *time.Location
}

// Validate checks if the configuration is valid
func (c *ReminderConfig) Validate() error {
	if c.TargetDatabaseID == "" {
		return &ValidationError{Field: "TargetDatabaseID", Message: "required"}
	}
	if len(c.ReminderTimings) == 0 {
		return &ValidationError{Field: "ReminderTimings", Message: "at least one timing required"}
	}
	if c.NotificationChannel == "" {
		return &ValidationError{Field: "NotificationChannel", Message: "required"}
	}
	if c.DatePropertyName == "" {
		c.DatePropertyName = "期限日" // Fixed
	}
	if c.TitlePropertyName == "" {
		c.TitlePropertyName = "タイトル" // Fixed
	}
	if c.Timezone == nil {
		c.Timezone = time.FixedZone("JST", 9*3600) // Fixed to JST
	}
	return nil
}
