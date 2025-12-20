package service

import (
	"context"
	"fmt"
	"schedule-reminder/internal/domain/calculator"
	"schedule-reminder/internal/domain/model"
	"schedule-reminder/internal/infrastructure/notifier"
	"strings"
	"time"
)

// NotionClient interface for Notion operations
type NotionClient interface {
	LoadReminderConfigs(ctx context.Context, masterDBID string) ([]*model.ReminderConfig, error)
	FetchSchedules(ctx context.Context, config *model.ReminderConfig, today time.Time) ([]*model.Schedule, error)
}

// ReminderService orchestrates the reminder processing logic
type ReminderService struct {
	notionClient NotionClient
	masterDBID   string
}

// NewReminderService creates a new reminder service
func NewReminderService(notionClient NotionClient, masterDBID string) *ReminderService {
	return &ReminderService{
		notionClient: notionClient,
		masterDBID:   masterDBID,
	}
}

// ProcessReminders is the main entry point for processing reminders
func (s *ReminderService) ProcessReminders(ctx context.Context) error {
	// Load all reminder configurations
	configs, err := s.notionClient.LoadReminderConfigs(ctx, s.masterDBID)
	if err != nil {
		return fmt.Errorf("failed to load configurations: %w", err)
	}

	fmt.Printf("Loaded %d reminder configurations\n", len(configs))

	// Process each configuration
	totalNotifications := 0
	for _, config := range configs {
		count, err := s.processConfig(ctx, config)
		if err != nil {
			fmt.Printf("Error processing config %s: %v\n", config.Name, err)
			continue // Continue with other configs even if one fails
		}
		totalNotifications += count
	}

	fmt.Printf("Sent %d notifications total\n", totalNotifications)
	return nil
}

// processConfig processes a single reminder configuration
func (s *ReminderService) processConfig(ctx context.Context, config *model.ReminderConfig) (int, error) {
	fmt.Printf("Processing: %s\n", config.Name)

	// Get today's date in the configured timezone
	today := time.Now().In(config.Timezone)

	// Fetch schedules from the target database
	schedules, err := s.notionClient.FetchSchedules(ctx, config, today)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch schedules: %w", err)
	}

	fmt.Printf("  Found %d schedules\n", len(schedules))

	// Create business day calculator
	holidays := loadHolidays() // TODO: Load from config or external source
	calc := calculator.NewBusinessDayCalculator(holidays, config.Timezone)

	// Process each schedule
	notificationCount := 0
	for _, schedule := range schedules {
		// Evaluate which timings should trigger today
		timings := s.evaluateTimings(schedule, config, today, calc)

		if len(timings) > 0 {
			fmt.Printf("    - %s (Due: %s) -> Timings: %v\n",
				schedule.Title,
				schedule.DueDate.Format("2006-01-02"),
				timings)

			// Send notifications for each triggered timing
			for _, timing := range timings {
				if err := s.sendNotification(ctx, schedule, config, timing); err != nil {
					fmt.Printf("      Error sending notification: %v\n", err)
					continue
				}
				notificationCount++
			}
		}
	}

	return notificationCount, nil
}

// evaluateTimings determines which reminder timings should trigger today
func (s *ReminderService) evaluateTimings(schedule *model.Schedule, config *model.ReminderConfig, today time.Time, calc *calculator.BusinessDayCalculator) []string {
	var triggered []string

	timings := config.ReminderTimings
	if len(schedule.ReminderTimings) > 0 {
		timings = schedule.ReminderTimings
	}

	for _, timing := range timings {
		reminderDate, err := calculator.ParseAndCalculateReminderDate(schedule.DueDate, timing, calc)
		if err != nil {
			fmt.Printf("      Warning: failed to calculate reminder date for '%s': %v\n", timing, err)
			continue
		}

		if calculator.IsSameDate(reminderDate, today) {
			triggered = append(triggered, timing)
		}
	}

	return triggered
}

// sendNotification sends a single notification
func (s *ReminderService) sendNotification(ctx context.Context, schedule *model.Schedule, config *model.ReminderConfig, timing string) error {
	// Build message from template
	message := BuildMessage(schedule, config, timing)

	destination := config.WebhookURL
	if strings.ToLower(config.NotificationChannel) == "line" {
		destination = config.LineRecipientID
	}

	// Create notification
	notification := &model.Notification{
		Schedule:    schedule,
		Config:      config,
		Timing:      timing,
		Message:     message,
		Destination: destination,
	}

	// Create notifier
	n, err := notifier.CreateNotifier(config)
	if err != nil {
		return fmt.Errorf("failed to create notifier: %w", err)
	}

	// Send notification
	if err := n.Send(ctx, notification); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	fmt.Printf("      ✓ Sent %s notification\n", n.Type())
	return nil
}

// loadHolidays loads holiday data
// TODO: Implement proper holiday loading (from API, config, etc.)
func loadHolidays() []time.Time {
	// For now, return empty list
	// In production, load from:
	// - Environment variable
	// - External API (e.g., Japanese holidays API)
	// - Configuration file
	return []time.Time{}
}
