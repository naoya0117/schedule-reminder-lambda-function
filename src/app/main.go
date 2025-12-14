package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"

	"schedule-reminder/internal/domain/service"
	awsinfra "schedule-reminder/internal/infrastructure/aws"
	"schedule-reminder/internal/infrastructure/notion"
)

// handler is the Lambda function handler for scheduled events
func handler(ctx context.Context) error {
	fmt.Println("=== Schedule Reminder Lambda Started ===")

	// Create SSM client to retrieve parameters from Parameter Store
	ssmClient, err := awsinfra.NewSSMClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create SSM client: %w", err)
	}

	// Get configuration from Parameter Store
	notionAPIKey, err := ssmClient.GetParameterWithFallback(ctx, "NOTION_API_KEY")
	if err != nil {
		return fmt.Errorf("failed to get NOTION_API_KEY: %w", err)
	}

	masterDBID, err := ssmClient.GetParameterWithFallback(ctx, "REMINDER_CONFIG_DB_ID")
	if err != nil {
		return fmt.Errorf("failed to get REMINDER_CONFIG_DB_ID: %w", err)
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
