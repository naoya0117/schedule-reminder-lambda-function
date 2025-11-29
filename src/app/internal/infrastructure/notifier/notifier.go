package notifier

import (
	"context"
	"schedule-reminder/internal/domain/model"
)

// Notifier is an interface for sending notifications
type Notifier interface {
	Send(ctx context.Context, notification *model.Notification) error
	Type() string
}
