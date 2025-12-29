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
	start := notionapi.Date(today)
	query := &notionapi.DatabaseQueryRequest{
		Filter: &notionapi.PropertyFilter{
			Property: config.DatePropertyName,
			Date: &notionapi.DateFilterCondition{
				OnOrAfter: &start,
			},
		},
		Sorts: []notionapi.SortObject{
			{
				Property:  config.DatePropertyName,
				Direction: notionapi.SortOrderASC,
			},
		},
	}

	var schedules []*model.Schedule
	for {
		result, err := c.client.Database.Query(ctx, notionapi.DatabaseID(config.TargetDatabaseID), query)
		if err != nil {
			return nil, fmt.Errorf("failed to query database %s: %w", config.TargetDatabaseID, err)
		}

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

		if !result.HasMore || result.NextCursor == "" {
			break
		}
		query.StartCursor = result.NextCursor
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
			schedule.DueDate = time.Time(*dp.Date.Start)
		}
	}

	// Extract description (optional)
	if descProp := getScheduleRichTextProperty(page, "説明", "Description"); descProp != nil && len(descProp.RichText) > 0 {
		schedule.Description = descProp.RichText[0].PlainText
	}
	// Extract reminder message (optional)
	if formulaProp := getScheduleFormulaProperty(page, "リマインドメッセージ"); formulaProp != nil {
		if value := formatFormulaValue(formulaProp.Formula); value != "" {
			schedule.MessageTemplate = value
		}
	} else if textProp := getScheduleRichTextProperty(page, "リマインドメッセージ"); textProp != nil && len(textProp.RichText) > 0 {
		schedule.MessageTemplate = textProp.RichText[0].PlainText
	}
	// Extract reminder timings (optional)
	if multiSelectProp := getScheduleMultiSelectProperty(page, "リマインドタイミング", "Reminder Timings"); multiSelectProp != nil {
		for _, option := range multiSelectProp.MultiSelect {
			schedule.ReminderTimings = append(schedule.ReminderTimings, option.Name)
		}
	}

	// Store all properties for template rendering
	for key, prop := range page.Properties {
		schedule.Properties[key] = extractPropertyValue(prop)
	}

	return schedule, nil
}

func getScheduleRichTextProperty(page notionapi.Page, names ...string) *notionapi.RichTextProperty {
	for _, name := range names {
		if prop, ok := page.Properties[name].(*notionapi.RichTextProperty); ok {
			return prop
		}
	}
	return nil
}

func getScheduleMultiSelectProperty(page notionapi.Page, names ...string) *notionapi.MultiSelectProperty {
	for _, name := range names {
		if prop, ok := page.Properties[name].(*notionapi.MultiSelectProperty); ok {
			return prop
		}
	}
	return nil
}

func getScheduleFormulaProperty(page notionapi.Page, names ...string) *notionapi.FormulaProperty {
	for _, name := range names {
		if prop, ok := page.Properties[name].(*notionapi.FormulaProperty); ok {
			return prop
		}
	}
	return nil
}

func formatFormulaValue(formula notionapi.Formula) string {
	switch formula.Type {
	case notionapi.FormulaTypeString:
		return formula.String
	case notionapi.FormulaTypeNumber:
		return fmt.Sprintf("%v", formula.Number)
	case notionapi.FormulaTypeBoolean:
		return fmt.Sprintf("%v", formula.Boolean)
	case notionapi.FormulaTypeDate:
		if formula.Date != nil && formula.Date.Start != nil {
			return time.Time(*formula.Date.Start).Format("2006-01-02")
		}
	}
	return ""
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
		if p.Select.Name != "" {
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
			return time.Time(*p.Date.Start).Format("2006-01-02")
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
