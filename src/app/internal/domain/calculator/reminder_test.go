package calculator

import (
	"testing"
	"time"
)

func TestParseAndCalculateReminderDate(t *testing.T) {
	loc := time.FixedZone("JST", 9*3600)
	due := time.Date(2024, 1, 8, 9, 0, 0, 0, loc) // Monday
	calc := NewBusinessDayCalculator(nil, loc)

	tests := []struct {
		name        string
		timing      string
		want        time.Time
		expectError bool
	}{
		{"same day", "当日", due, false},
		{"days before", "3日前", due.AddDate(0, 0, -3), false},
		{"weeks before", "2週間前", due.AddDate(0, 0, -14), false},
		{"business days before across weekend", "2営業日前", time.Date(2024, 1, 4, 9, 0, 0, 0, loc), false},
		{"unsupported format", "invalid", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAndCalculateReminderDate(due, tt.timing, calc)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error for %q", tt.timing)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("timing %q: got %s, want %s", tt.timing, got, tt.want)
			}
		})
	}
}

func TestFormatDaysText(t *testing.T) {
	tests := []struct {
		timing string
		want   string
	}{
		{"当日", "今日"},
		{"1日前", "明日"},
		{"2日前", "2日後"},
		{"3週間前", "3日後"}, // fallback to a number prefix
	}

	for _, tt := range tests {
		if got := FormatDaysText(tt.timing); got != tt.want {
			t.Fatalf("timing %q: got %q, want %q", tt.timing, got, tt.want)
		}
	}
}
