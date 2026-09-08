// Package orchestrator ties the assignment engine, Codeforces activity
// checks, and Telegram notifications into one pipeline run. It depends
// only on small local interfaces for Sheets/Codeforces/Telegram, never
// on their concrete packages — internal/sheets.Client,
// internal/Codeforces.Client, and internal/telegram.Client each already
// satisfy the corresponding interface here without any changes, so
// production wiring is a one-line assignment (see cmd/sync). Depending
// on interfaces instead of the concrete clients also means this whole
// pipeline can be exercised in tests with fakes, with no live
// credentials or network access required.
package orchestrator

import (
	"context"
	"fmt"
	"time"

	"fcds-training-tool/internal/Codeforces"
	"fcds-training-tool/internal/assignment"
	"fcds-training-tool/internal/policy"
	"fcds-training-tool/internal/roster"
)

type SheetsClient interface {
	Read(ctx context.Context, spreadsheetID, readRange string) ([][]string, error)
	Write(ctx context.Context, spreadsheetID, writeRange string, rows [][]string) error
}

type CodeforcesClient interface {
	UserStatus(ctx context.Context, handle string, count int) ([]codeforces.Submission, error)
}

type Notifier interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

// Config bundles everything one Run needs: where the data lives, and
// the thresholds/windows that drive decisions.
type Config struct {
	SpreadsheetID      string
	TrainersRange      string
	TraineesRange      string
	AssignmentsRange   string
	Policy             policy.Config
	ActivityWindow     time.Duration // how far back "recent activity" looks
	SubmissionsToFetch int           // how many recent submissions to pull per handle
}

// Orchestrator runs one full sync cycle against the given clients.
type Orchestrator struct {
	Sheets     SheetsClient
	Codeforces CodeforcesClient
	Notifier   Notifier
	Config     Config
	// Now is overridable so tests can fix "the current time" instead
	// of racing a real clock; it defaults to time.Now.
	Now func() time.Time
}

// RunResult summarizes what one Run did, for logging and for tests.
type RunResult struct {
	Warned       []string
	Filtered     []string
	Promoted     []string
	Assignment   assignment.Assignment
	Notified     int
	NotifyErrors []string
}

func (o *Orchestrator) now() time.Time {
	if o.Now != nil {
		return o.Now()
	}
	return time.Now()
}

// Run performs one full cycle:
//  1. read trainers/trainees/previous-assignment from the sheet
//  2. fetch each trainee's Codeforces activity and evaluate the policy
//  3. apply warnings, filtering, and waitlist promotions
//  4. compute a fresh trainer/trainee assignment for everyone now active
//  5. write the updated roster and assignment back
//  6. notify trainees whose status or trainer actually changed
func (o *Orchestrator) Run(ctx context.Context) (RunResult, error) {
	var result RunResult

	trainerRows, err := o.Sheets.Read(ctx, o.Config.SpreadsheetID, o.Config.TrainersRange)
	if err != nil {
		return result, fmt.Errorf("read trainers: %w", err)
	}
	trainers, err := roster.ParseTrainers(trainerRows)
	if err != nil {
		return result, fmt.Errorf("parse trainers: %w", err)
	}

	traineeRows, err := o.Sheets.Read(ctx, o.Config.SpreadsheetID, o.Config.TraineesRange)
	if err != nil {
		return result, fmt.Errorf("read trainees: %w", err)
	}
	trainees, err := roster.ParseTrainees(traineeRows)
	if err != nil {
		return result, fmt.Errorf("parse trainees: %w", err)
	}

	previousRows, err := o.Sheets.Read(ctx, o.Config.SpreadsheetID, o.Config.AssignmentsRange)
	if err != nil {
		return result, fmt.Errorf("read previous assignment: %w", err)
	}
	previousAssignment, err := roster.ParseAssignments(previousRows)
	if err != nil {
		return result, fmt.Errorf("parse previous assignment: %w", err)
	}

	now := o.now()
	for i, t := range trainees {
		if t.CodeforcesHandle == "" || t.Status == roster.StatusFiltered {
			continue // nothing to evaluate without a handle, or already filtered
		}
		submissions, err := o.Codeforces.UserStatus(ctx, t.CodeforcesHandle, o.Config.SubmissionsToFetch)
		if err != nil {
			return result, fmt.Errorf("fetch Codeforces activity for %s (%s): %w", t.ID, t.CodeforcesHandle, err)
		}
		activity := codeforces.Summarize(t.CodeforcesHandle, submissions, now, o.Config.ActivityWindow)
		decision := policy.Evaluate(o.Config.Policy, t, activity)

		switch decision.Action {
		case policy.ActionWarn:
			trainees[i].WarningCount++
			result.Warned = append(result.Warned, t.ID)
			o.notify(ctx, &result, t.TelegramChatID,
				fmt.Sprintf("Hi %s — %s. Keep solving to stay active in the program!", t.Name, decision.Reason))

		case policy.ActionFilter:
			trainees[i].Status = roster.StatusFiltered
			result.Filtered = append(result.Filtered, t.ID)
			o.notify(ctx, &result, t.TelegramChatID,
				fmt.Sprintf("Hi %s, you've been moved out of active training (%s). Please don't message your trainer going forward — reach out to the community if you'd like to rejoin later.", t.Name, decision.Reason))

		case policy.ActionPromote:
			trainees[i].Status = roster.StatusActive
			trainees[i].WarningCount = 0
			result.Promoted = append(result.Promoted, t.ID)
			// The "your trainer is now X" notification below covers
			// promoted trainees too, since they're brand new in the
			// assignment map.

		case policy.ActionNone:
			if trainees[i].WarningCount > 0 {
				trainees[i].WarningCount = 0 // back on track — clear the streak
			}
		}
	}

	newAssignment := assignment.Assign(trainers, roster.ActiveAssignmentInput(trainees))

	for traineeID, trainerID := range newAssignment.TraineeToTrainer {
		if previousAssignment[traineeID] == trainerID {
			continue // unchanged since last run, don't re-notify
		}
		trainee := traineeByID(trainees, traineeID)
		trainerName := trainerNameByID(trainers, trainerID)
		if trainee == nil {
			continue
		}
		o.notify(ctx, &result, trainee.TelegramChatID,
			fmt.Sprintf("Hi %s! Your trainer is now %s.", trainee.Name, trainerName))
	}

	if err := o.Sheets.Write(ctx, o.Config.SpreadsheetID, o.Config.TraineesRange, roster.RosterRows(trainees)); err != nil {
		return result, fmt.Errorf("write updated roster: %w", err)
	}
	if err := o.Sheets.Write(ctx, o.Config.SpreadsheetID, o.Config.AssignmentsRange, roster.AssignmentRows(newAssignment.TraineeToTrainer)); err != nil {
		return result, fmt.Errorf("write updated assignment: %w", err)
	}

	result.Assignment = newAssignment
	return result, nil
}

func (o *Orchestrator) notify(ctx context.Context, result *RunResult, chatID int64, text string) {
	if chatID == 0 {
		return // no linked Telegram account for this trainee
	}
	if err := o.Notifier.SendMessage(ctx, chatID, text); err != nil {
		result.NotifyErrors = append(result.NotifyErrors, fmt.Sprintf("chat %d: %v", chatID, err))
		return
	}
	result.Notified++
}

func traineeByID(trainees []roster.Trainee, id string) *roster.Trainee {
	for i := range trainees {
		if trainees[i].ID == id {
			return &trainees[i]
		}
	}
	return nil
}

func trainerNameByID(trainers []assignment.Trainer, id string) string {
	for _, t := range trainers {
		if t.ID == id {
			return t.Name
		}
	}
	return id
}
