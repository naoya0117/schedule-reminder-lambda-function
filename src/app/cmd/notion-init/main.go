package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jomei/notionapi"
)

type options struct {
	apiKey                string
	parentPageID          string
	configDBName          string
	scheduleDBName        string
	skipScheduleDB        bool
	createSampleConfig    bool
	reminderTimingOptions []string
	notificationChannels  []string
	sampleConfigName      string
	sampleReminderTimings []string
	sampleNotification    string
	sampleWebhookURL      string
	sampleChannelToken    string
}

func main() {
	opts := parseFlags()
	if err := validateOptions(opts); err != nil {
		log.Fatalf("invalid options: %v", err)
	}

	ctx := context.Background()
	client := notionapi.NewClient(notionapi.Token(opts.apiKey))

	masterDB, err := client.Database.Create(ctx, buildMasterDatabaseRequest(opts))
	if err != nil {
		log.Fatalf("failed to create master database: %v", err)
	}

	var scheduleDB *notionapi.Database
	if !opts.skipScheduleDB {
		scheduleDB, err = client.Database.Create(ctx, buildScheduleDatabaseRequest(opts))
		if err != nil {
			log.Fatalf("failed to create schedule database: %v", err)
		}
	}

	var samplePage *notionapi.Page
	if opts.createSampleConfig {
		if scheduleDB == nil {
			log.Fatal("sample config requires schedule database (omit --skip-schedule-db)")
		}
		samplePage, err = client.Page.Create(ctx, &notionapi.PageCreateRequest{
			Parent: notionapi.Parent{
				Type:       notionapi.ParentTypeDatabaseID,
				DatabaseID: notionapi.DatabaseID(masterDB.ID),
			},
			Properties: buildSampleConfigProperties(opts, scheduleDB.ID.String()),
		})
		if err != nil {
			log.Fatalf("failed to create sample config page: %v", err)
		}
	}

	printResult(masterDB, scheduleDB, samplePage)
}

func parseFlags() options {
	var opts options

	flag.StringVar(&opts.apiKey, "api-key", "", "Notion API key (or NOTION_API_KEY)")
	flag.StringVar(&opts.parentPageID, "parent-page-id", "", "Parent Notion page ID (or NOTION_PARENT_PAGE_ID)")
	flag.StringVar(&opts.configDBName, "config-db-name", "リマインダー設定マスター", "Master config database name")
	flag.StringVar(&opts.scheduleDBName, "schedule-db-name", "スケジュール", "Schedule database name")
	flag.BoolVar(&opts.skipScheduleDB, "skip-schedule-db", false, "Skip creating the schedule database")
	flag.BoolVar(&opts.createSampleConfig, "create-sample-config", false, "Create a sample config row in the master database")
	flag.StringVar(&opts.sampleConfigName, "sample-config-name", "サンプルリマインド", "Sample config name")
	flag.StringVar(&opts.sampleNotification, "sample-notification-channel", "Discord", "Sample notification channel")
	flag.StringVar(&opts.sampleWebhookURL, "sample-webhook-url", "", "Sample webhook URL (required when creating sample config)")
	flag.StringVar(&opts.sampleChannelToken, "sample-channel-token", "", "Sample channel access token")

	reminderTimingOptions := flag.String("reminder-timing-options", "当日,1日前,2日前,3日前,1営業日前,2営業日前,3営業日前,4営業日前,5営業日前,1週間前,2週間前", "Comma-separated reminder timing options")
	notificationChannels := flag.String("notification-channels", "Discord,LINE,Slack", "Comma-separated notification channels")
	sampleReminderTimings := flag.String("sample-reminder-timings", "当日,1日前", "Comma-separated reminder timings for sample config")

	flag.Parse()

	if opts.apiKey == "" {
		opts.apiKey = firstEnvValue("NOTION_API_KEY", "NOTION_TOKEN")
	}
	if opts.parentPageID == "" {
		opts.parentPageID = firstEnvValue("NOTION_PARENT_PAGE_ID", "NOTION_PARENT_PAGE")
	}

	opts.reminderTimingOptions = splitList(*reminderTimingOptions)
	opts.notificationChannels = splitList(*notificationChannels)
	opts.sampleReminderTimings = splitList(*sampleReminderTimings)

	return opts
}

func validateOptions(opts options) error {
	if opts.apiKey == "" {
		return fmt.Errorf("api-key is required")
	}
	if opts.parentPageID == "" {
		return fmt.Errorf("parent-page-id is required")
	}
	if len(opts.reminderTimingOptions) == 0 {
		return fmt.Errorf("reminder-timing-options must not be empty")
	}
	if len(opts.notificationChannels) == 0 {
		return fmt.Errorf("notification-channels must not be empty")
	}
	if opts.createSampleConfig && opts.sampleWebhookURL == "" {
		return fmt.Errorf("sample-webhook-url is required when create-sample-config is set")
	}
	if opts.createSampleConfig && opts.skipScheduleDB {
		return fmt.Errorf("create-sample-config requires schedule database creation")
	}
	return nil
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func firstEnvValue(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func buildMasterDatabaseRequest(opts options) *notionapi.DatabaseCreateRequest {
	return &notionapi.DatabaseCreateRequest{
		Parent: notionapi.Parent{
			Type:   notionapi.ParentTypePageID,
			PageID: notionapi.PageID(opts.parentPageID),
		},
		Title:      []notionapi.RichText{richText(opts.configDBName)},
		IsInline:   true,
		Properties: masterDatabaseProperties(opts),
	}
}

func masterDatabaseProperties(opts options) notionapi.PropertyConfigs {
	return notionapi.PropertyConfigs{
		"名前": &notionapi.TitlePropertyConfig{
			Type: notionapi.PropertyConfigTypeTitle,
		},
		"有効": &notionapi.CheckboxPropertyConfig{
			Type: notionapi.PropertyConfigTypeCheckbox,
		},
		"対象データベースID": &notionapi.RichTextPropertyConfig{
			Type: notionapi.PropertyConfigTypeRichText,
		},
		"リマインドタイミング": &notionapi.MultiSelectPropertyConfig{
			Type:        notionapi.PropertyConfigTypeMultiSelect,
			MultiSelect: notionapi.Select{Options: toOptions(opts.reminderTimingOptions)},
		},
		"通知チャネル": &notionapi.SelectPropertyConfig{
			Type:   notionapi.PropertyConfigTypeSelect,
			Select: notionapi.Select{Options: toOptions(opts.notificationChannels)},
		},
		"Webhook URL": &notionapi.URLPropertyConfig{
			Type: notionapi.PropertyConfigTypeURL,
		},
		"チャネルアクセストークン": &notionapi.RichTextPropertyConfig{
			Type: notionapi.PropertyConfigTypeRichText,
		},
	}
}

func buildScheduleDatabaseRequest(opts options) *notionapi.DatabaseCreateRequest {
	return &notionapi.DatabaseCreateRequest{
		Parent: notionapi.Parent{
			Type:   notionapi.ParentTypePageID,
			PageID: notionapi.PageID(opts.parentPageID),
		},
		Title:    []notionapi.RichText{richText(opts.scheduleDBName)},
		IsInline: true,
		Properties: notionapi.PropertyConfigs{
			"タイトル": &notionapi.TitlePropertyConfig{
				Type: notionapi.PropertyConfigTypeTitle,
			},
			"期限日": &notionapi.DatePropertyConfig{
				Type: notionapi.PropertyConfigTypeDate,
			},
			"説明": &notionapi.RichTextPropertyConfig{
				Type: notionapi.PropertyConfigTypeRichText,
			},
			"リマインドタイミング": &notionapi.MultiSelectPropertyConfig{
				Type:        notionapi.PropertyConfigTypeMultiSelect,
				MultiSelect: notionapi.Select{Options: toOptions(opts.reminderTimingOptions)},
			},
			"メッセージテンプレート": &notionapi.RichTextPropertyConfig{
				Type: notionapi.PropertyConfigTypeRichText,
			},
		},
	}
}

func buildSampleConfigProperties(opts options, scheduleDBID string) notionapi.Properties {
	props := notionapi.Properties{
		"名前": &notionapi.TitleProperty{
			Type:  notionapi.PropertyTypeTitle,
			Title: []notionapi.RichText{richText(opts.sampleConfigName)},
		},
		"有効": &notionapi.CheckboxProperty{
			Type:     notionapi.PropertyTypeCheckbox,
			Checkbox: true,
		},
		"対象データベースID": richTextProperty(scheduleDBID),
		"リマインドタイミング": &notionapi.MultiSelectProperty{
			Type:        notionapi.PropertyTypeMultiSelect,
			MultiSelect: toOptions(opts.sampleReminderTimings),
		},
		"通知チャネル": &notionapi.SelectProperty{
			Type:   notionapi.PropertyTypeSelect,
			Select: notionapi.Option{Name: opts.sampleNotification},
		},
		"Webhook URL": &notionapi.URLProperty{
			Type: notionapi.PropertyTypeURL,
			URL:  opts.sampleWebhookURL,
		},
	}

	if opts.sampleChannelToken != "" {
		props["チャネルアクセストークン"] = richTextProperty(opts.sampleChannelToken)
	}

	return props
}

func richTextProperty(value string) *notionapi.RichTextProperty {
	return &notionapi.RichTextProperty{
		Type:     notionapi.PropertyTypeRichText,
		RichText: []notionapi.RichText{richText(value)},
	}
}

func richText(value string) notionapi.RichText {
	return notionapi.RichText{
		Type: notionapi.ObjectTypeText,
		Text: &notionapi.Text{Content: value},
	}
}

func toOptions(values []string) []notionapi.Option {
	options := make([]notionapi.Option, 0, len(values))
	for _, value := range values {
		options = append(options, notionapi.Option{Name: value})
	}
	return options
}

func printResult(masterDB *notionapi.Database, scheduleDB *notionapi.Database, samplePage *notionapi.Page) {
	fmt.Println("Notion databases initialized.")
	fmt.Printf("Master DB ID: %s\n", masterDB.ID.String())
	fmt.Printf("Master DB URL: %s\n", masterDB.URL)

	if scheduleDB != nil {
		fmt.Printf("Schedule DB ID: %s\n", scheduleDB.ID.String())
		fmt.Printf("Schedule DB URL: %s\n", scheduleDB.URL)
	}
	if samplePage != nil {
		fmt.Printf("Sample Config Page ID: %s\n", samplePage.ID.String())
		fmt.Printf("Sample Config Page URL: %s\n", samplePage.URL)
	}
}
