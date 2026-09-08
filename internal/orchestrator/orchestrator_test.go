package orchestrator

import (
	"context"
	"testing"
	"time"

	"fcds-training-tool/internal/Codeforces"
	"fcds-training-tool/internal/policy"
)

// fakeSheets is an in-memory stand-in for internal/sheets.Client: no
// credentials, no network, just a map keyed by range.
type fakeSheets struct {
	data map[string][][]string
}

func (f *fakeSheets) Read(_ context.Context, _, readRange string) ([][]string, error) {
	return f.data[readRange], nil
}

func (f *fakeSheets) Write(_ context.Context, _, writeRange string, rows [][]string) error {
	if f.data == nil {
		f.data = make(map[string][][]string)
	}
	f.data[writeRange] = rows
	return nil
}

// fakeCodeforces returns a canned submission list per handle instead
// of calling the real API.
type fakeCodeforces struct {
	byHandle map[string][]codeforces.Submission
}

func (f *fakeCodeforces) UserStatus(_ context.Context, handle string, _ int) ([]codeforces.Submission, error) {
	return f.byHandle[handle], nil
}

// fakeNotifier records every message instead of calling Telegram.
type fakeNotifier struct {
	sent []string // "chatID:text"
}

func (f *fakeNotifier) SendMessage(_ context.Context, chatID int64, text string) error {
	f.sent = append(f.sent, itoa(chatID)+":"+text)
	return nil
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func accepted(problemIndex string, age time.Duration, now time.Time) codeforces.Submission {
	return codeforces.Submission{
		Problem:             codeforces.Problem{ContestID: 1, Index: problemIndex},
		Verdict:             "OK",
		CreationTimeSeconds: now.Add(-age).Unix(),
	}
}

func TestOrchestrator_FullCycle(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	weekAgo := 8 * 24 * time.Hour // just outside a 7-day window

	trainersCSV := [][]string{
		{"ID", "Name", "Gender"},
		{"tr-sara", "Sara", "female"},
		{"tr-ahmed", "Ahmed", "male"},
	}
	traineesCSV := [][]string{
		{"ID", "Name", "Gender", "CodeforcesHandle", "Status", "WarningCount", "TelegramChatID"},
		{"s1", "Youssef", "male", "s1cf", "active", "0", "100"},   // low activity -> warn
		{"s2", "Laila", "female", "s2cf", "active", "1", "200"},   // low activity, already warned once -> filter
		{"s3", "Karim", "male", "s3cf", "waitlisted", "0", "300"}, // high activity -> promote
		{"s4", "Nour", "female", "s4cf", "active", "0", "400"},    // meets threshold -> stays, no warning
		{"s5", "Hassan", "male", "", "active", "0", "0"},          // no CF handle -> skipped entirely
	}
	previousAssignmentCSV := [][]string{
		{"Trainee ID", "Trainer ID"},
		{"s1", "tr-ahmed"},
		{"s4", "tr-sara"},
	}

	sheets := &fakeSheets{data: map[string][][]string{
		"Trainers!A:C":    trainersCSV,
		"Trainees!A:G":    traineesCSV,
		"Assignments!A:B": previousAssignmentCSV,
	}}
	cf := &fakeCodeforces{byHandle: map[string][]codeforces.Submission{
		"s1cf": {},                            // no submissions at all -> 0 solved in period
		"s2cf": {accepted("A", weekAgo, now)}, // solved, but outside the activity window
		"s3cf": {
			accepted("A", time.Hour, now),
			accepted("B", 2*time.Hour, now),
			accepted("C", 3*time.Hour, now),
		},
		"s4cf": {accepted("A", time.Hour, now)},
	}}
	notifier := &fakeNotifier{}

	o := &Orchestrator{
		Sheets:     sheets,
		Codeforces: cf,
		Notifier:   notifier,
		Now:        func() time.Time { return now },
		Config: Config{
			SpreadsheetID:      "sheet-1",
			TrainersRange:      "Trainers!A:C",
			TraineesRange:      "Trainees!A:G",
			AssignmentsRange:   "Assignments!A:B",
			Policy:             policy.DefaultConfig(),
			ActivityWindow:     7 * 24 * time.Hour,
			SubmissionsToFetch: 50,
		},
	}

	result, err := o.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(result.Warned) != 1 || result.Warned[0] != "s1" {
		t.Errorf("Warned = %v, want [s1]", result.Warned)
	}
	if len(result.Filtered) != 1 || result.Filtered[0] != "s2" {
		t.Errorf("Filtered = %v, want [s2]", result.Filtered)
	}
	if len(result.Promoted) != 1 || result.Promoted[0] != "s3" {
		t.Errorf("Promoted = %v, want [s3]", result.Promoted)
	}
	if len(result.Assignment.Unassigned) != 0 {
		t.Errorf("Unassigned = %v, want none", result.Assignment.Unassigned)
	}
	// s2 was filtered, so she must not appear in the new assignment.
	if _, ok := result.Assignment.TraineeToTrainer["s2"]; ok {
		t.Errorf("filtered trainee s2 should not appear in the new assignment")
	}
	// s5 has no CF handle and was never evaluated, so she stays active
	// and must still be assignable.
	if _, ok := result.Assignment.TraineeToTrainer["s5"]; !ok {
		t.Errorf("s5 (no CF handle, still active) should still be in the new assignment")
	}

	// s4's trainer didn't change (tr-sara before and after) -> no notification.
	// s1's trainer also didn't change (tr-ahmed) -> only the warning notification, not an assignment one.
	// s3 is brand new (promoted) -> must get a "your trainer is now" message.
	foundPromotionNotice := false
	for _, msg := range notifier.sent {
		if contains(msg, "300:") && contains(msg, "Your trainer is now") {
			foundPromotionNotice = true
		}
	}
	if !foundPromotionNotice {
		t.Errorf("expected a trainer-assignment notice sent to s3 (chat 300), got: %v", notifier.sent)
	}

	// Roster and assignment sheets should have been written back.
	if len(sheets.data["Trainees!A:G"]) != len(traineesCSV) {
		t.Errorf("expected the roster sheet to be rewritten with the same row count")
	}
	if _, ok := sheets.data["Assignments!A:B"]; !ok {
		t.Errorf("expected the assignments sheet to be rewritten")
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestOrchestrator_SkipsTraineesWithoutCodeforcesHandle(t *testing.T) {
	now := time.Now()
	sheets := &fakeSheets{data: map[string][][]string{
		"Trainers!A:C": {{"ID", "Name", "Gender"}, {"t1", "Ahmed", "male"}},
		"Trainees!A:G": {
			{"ID", "Name", "Gender", "CodeforcesHandle", "Status", "WarningCount", "TelegramChatID"},
			{"s1", "Hassan", "male", "", "active", "0", "0"},
		},
		"Assignments!A:B": {{"Trainee ID", "Trainer ID"}},
	}}
	cf := &fakeCodeforces{byHandle: map[string][]codeforces.Submission{}}
	notifier := &fakeNotifier{}

	o := &Orchestrator{
		Sheets: sheets, Codeforces: cf, Notifier: notifier,
		Now: func() time.Time { return now },
		Config: Config{
			SpreadsheetID: "s", TrainersRange: "Trainers!A:C", TraineesRange: "Trainees!A:G",
			AssignmentsRange: "Assignments!A:B", Policy: policy.DefaultConfig(),
			ActivityWindow: 7 * 24 * time.Hour, SubmissionsToFetch: 50,
		},
	}
	result, err := o.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(result.Warned) != 0 || len(result.Filtered) != 0 {
		t.Errorf("a trainee with no CF handle should never be warned or filtered, got warned=%v filtered=%v", result.Warned, result.Filtered)
	}
	if result.Assignment.TraineeToTrainer["s1"] != "t1" {
		t.Errorf("expected s1 still assigned to t1")
	}
}
