// Command sync runs one full orchestrator cycle against the real
// Google Sheet, the real Codeforces API, and the real Telegram bot.
// Configuration comes entirely from environment variables (see
// config.example.env) so no credentials ever need to be hardcoded or
// committed.
//
// This package imports internal/sheets, which needs
// google.golang.org/api — run `go get google.golang.org/api/sheets/v4
// && go mod tidy` once before building this if go.mod doesn't already
// list it.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"fcds-training-tool/internal/Codeforces"
	"fcds-training-tool/internal/orchestrator"
	"fcds-training-tool/internal/policy"
	"fcds-training-tool/internal/sheets"
	"fcds-training-tool/internal/telegram"
)

func requireEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("missing required environment variable %s (see config.example.env)", name)
	}
	return value
}

func main() {
	ctx := context.Background()

	credentialsFile := requireEnv("GOOGLE_CREDENTIALS_FILE")
	spreadsheetID := requireEnv("GOOGLE_SPREADSHEET_ID")
	trainersRange := envOr("GOOGLE_TRAINERS_RANGE", "Trainers!A:C")
	traineesRange := envOr("GOOGLE_TRAINEES_RANGE", "Trainees!A:G")
	assignmentsRange := envOr("GOOGLE_ASSIGNMENTS_RANGE", "Assignments!A:B")
	botToken := requireEnv("TELEGRAM_BOT_TOKEN")

	sheetsClient, err := sheets.NewClient(ctx, credentialsFile)
	if err != nil {
		log.Fatalf("create Google Sheets client: %v", err)
	}
	cfClient := codeforces.NewClient(nil)
	telegramClient := telegram.NewClient(botToken, nil)

	o := &orchestrator.Orchestrator{
		Sheets:     sheetsClient,
		Codeforces: cfClient,
		Notifier:   telegramClient,
		Config: orchestrator.Config{
			SpreadsheetID:      spreadsheetID,
			TrainersRange:      trainersRange,
			TraineesRange:      traineesRange,
			AssignmentsRange:   assignmentsRange,
			Policy:             policy.DefaultConfig(),
			ActivityWindow:     7 * 24 * time.Hour,
			SubmissionsToFetch: 100,
		},
	}

	result, err := o.Run(ctx)
	if err != nil {
		log.Fatalf("sync run failed: %v", err)
	}

	log.Printf("sync complete: %d warned, %d filtered, %d promoted, %d notifications sent, %d unassigned",
		len(result.Warned), len(result.Filtered), len(result.Promoted), result.Notified, len(result.Assignment.Unassigned))
	if len(result.NotifyErrors) > 0 {
		log.Printf("some notifications failed: %v", result.NotifyErrors)
	}
	if len(result.Assignment.Unassigned) > 0 {
		log.Printf("could not place: %v", result.Assignment.Unassigned)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
