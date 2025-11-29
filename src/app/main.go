package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"

	"schedule-reminder/internal/domain/service"
	"schedule-reminder/internal/infrastructure/notion"
)

// handler is the Lambda function handler for scheduled events
func handler(ctx context.Context) error {
	fmt.Println("=== Schedule Reminder Lambda Started ===")

	// Get configuration from environment variables
	notionAPIKey := os.Getenv("NOTION_API_KEY")
	if notionAPIKey == "" {
		return fmt.Errorf("NOTION_API_KEY environment variable is required")
	}

	masterDBID := os.Getenv("REMINDER_CONFIG_DB_ID")
	if masterDBID == "" {
		return fmt.Errorf("REMINDER_CONFIG_DB_ID environment variable is required")
	}

	// Create Notion client
	notionClient := notion.NewClient(notionAPIKey)

	// Create reminder service
	reminderService := service.NewReminderService(notionClient, masterDBID)

	// Process reminders
	if err := reminderService.ProcessReminders(ctx); err != nil {
		fmt.Printf("Error processing reminders: %v\n", err)
		return err
	}

	fmt.Println("=== Schedule Reminder Lambda Completed ===")
	return nil
}

func main() {
	lambda.Start(handler)
}
