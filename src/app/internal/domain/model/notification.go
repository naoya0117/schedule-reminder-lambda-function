package model

// Notification represents a notification to be sent
type Notification struct {
	Schedule   *Schedule
	Config     *ReminderConfig
	Timing     string
	Message    string
	Destination string
}
