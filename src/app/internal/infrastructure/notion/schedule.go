package notion

import (
	"context"
	"fmt"
	"schedule-reminder/internal/domain/model"
	"time"

	"github.com/jomei/notionapi"
)

// FetchSchedules fetches schedules from a child database
func (c *Client) FetchSchedules(ctx context.Context, config *model.ReminderConfig, today time.Time) ([]*model.Schedule, error) {
	// Query only future schedules (optimization)
	query := &notionapi.DatabaseQueryRequest{
		Filter: &notionapi.PropertyFilter{
			Property: config.DatePropertyName,
			Date: &notionapi.DateFilterCondition{
				OnOrAfter: notionapi.Date(today),
			},
		},
		Sorts: []notionapi.SortObject{
			{
				Property:  config.DatePropertyName,
				Direction: notionapi.SortOrderASC,
			},
		},
	}

	result, err := c.client.Database.Query(ctx, notionapi.DatabaseID(config.TargetDatabaseID), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database %s: %w", config.TargetDatabaseID, err)
	}

	var schedules []*model.Schedule
	for _, page := range result.Results {
		schedule, err := c.parseSchedule(page, config)
		if err != nil {
			fmt.Printf("Warning: failed to parse schedule %s: %v\n", page.ID, err)
			continue
		}

		if err := schedule.Validate(); err != nil {
			fmt.Printf("Warning: invalid schedule %s: %v\n", page.ID, err)
			continue
		}

		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// parseSchedule extracts schedule information from a Notion page
func (c *Client) parseSchedule(page notionapi.Page, config *model.ReminderConfig) (*model.Schedule, error) {
	schedule := &model.Schedule{
		ID:         page.ID.String(),
		NotionURL:  page.URL,
		Properties: make(map[string]interface{}),
	}

	// Extract title
	titleProp := page.Properties[config.TitlePropertyName]
	if titleProp != nil {
		if tp, ok := titleProp.(*notionapi.TitleProperty); ok && len(tp.Title) > 0 {
			schedule.Title = tp.Title[0].PlainText
		}
	}

	// Extract due date
	dateProp := page.Properties[config.DatePropertyName]
	if dateProp != nil {
		if dp, ok := dateProp.(*notionapi.DateProperty); ok && dp.Date != nil && dp.Date.Start != nil {
			schedule.DueDate = dp.Date.Start.Time
		}
	}

	// Extract description (optional)
	if descProp, ok := page.Properties["Description"].(*notionapi.RichTextProperty); ok && len(descProp.RichText) > 0 {
		schedule.Description = descProp.RichText[0].PlainText
	}

	// Store all properties for template rendering
	for key, prop := range page.Properties {
		schedule.Properties[key] = extractPropertyValue(prop)
	}

	return schedule, nil
}

// extractPropertyValue extracts the value from a Notion property
func extractPropertyValue(prop notionapi.Property) interface{} {
	switch p := prop.(type) {
	case *notionapi.TitleProperty:
		if len(p.Title) > 0 {
			return p.Title[0].PlainText
		}
	case *notionapi.RichTextProperty:
		if len(p.RichText) > 0 {
			return p.RichText[0].PlainText
		}
	case *notionapi.NumberProperty:
		return p.Number
	case *notionapi.SelectProperty:
		if p.Select != nil {
			return p.Select.Name
		}
	case *notionapi.MultiSelectProperty:
		var values []string
		for _, opt := range p.MultiSelect {
			values = append(values, opt.Name)
		}
		return values
	case *notionapi.DateProperty:
		if p.Date != nil && p.Date.Start != nil {
			return p.Date.Start.Time.Format("2006-01-02")
		}
	case *notionapi.PeopleProperty:
		var names []string
		for _, person := range p.People {
			if person.Name != "" {
				names = append(names, person.Name)
			}
		}
		return names
	case *notionapi.CheckboxProperty:
		return p.Checkbox
	case *notionapi.URLProperty:
		return string(p.URL)
	}
	return nil
}
