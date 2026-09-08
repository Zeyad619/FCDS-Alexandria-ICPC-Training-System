package roster

import (
	"fcds-training-tool/internal/assignment"
	"testing"
)

func TestParseTrainers(t *testing.T) {
	rows := [][]string{
		{"ID", "Name", "Gender"},
		{"t1", "Ahmed", "male"},
		{"t2", "Sara", "Female"},
		{"", "", ""}, // trailing blank row should be skipped
	}
	trainers, err := ParseTrainers(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(trainers) != 2 {
		t.Fatalf("got %d trainers, want 2", len(trainers))
	}
	if trainers[0].ID != "t1" || trainers[0].Gender != assignment.Male {
		t.Errorf("trainer 0 = %+v", trainers[0])
	}
	if trainers[1].ID != "t2" || trainers[1].Gender != assignment.Female {
		t.Errorf("trainer 1 = %+v", trainers[1])
	}
}

func TestParseTrainers_BadGender(t *testing.T) {
	rows := [][]string{
		{"ID", "Name", "Gender"},
		{"t1", "Ahmed", "unspecified"},
	}
	if _, err := ParseTrainers(rows); err == nil {
		t.Fatal("expected an error for an unrecognized gender, got nil")
	}
}

func TestParseTrainees_ColumnOrderIndependent(t *testing.T) {
	// Columns deliberately out of the "expected" order, to prove
	// parsing is driven by header names, not position.
	rows := [][]string{
		{"Status", "ID", "TelegramChatID", "Gender", "CodeforcesHandle", "Name", "WarningCount"},
		{"active", "s1", "555", "male", "youssef_cf", "Youssef", "1"},
		{"waitlisted", "s2", "", "female", "mona_cf", "Mona", ""},
	}
	trainees, err := ParseTrainees(rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(trainees) != 2 {
		t.Fatalf("got %d trainees, want 2", len(trainees))
	}
	if trainees[0].ID != "s1" || trainees[0].Status != StatusActive || trainees[0].WarningCount != 1 || trainees[0].TelegramChatID != 555 {
		t.Errorf("trainee 0 = %+v", trainees[0])
	}
	if trainees[1].ID != "s2" || trainees[1].Status != StatusWaitlisted || trainees[1].TelegramChatID != 0 {
		t.Errorf("trainee 1 = %+v", trainees[1])
	}
}

func TestParseTrainees_BadStatus(t *testing.T) {
	rows := [][]string{
		{"ID", "Name", "Gender", "CodeforcesHandle", "Status"},
		{"s1", "Youssef", "male", "youssef_cf", "graduated"},
	}
	if _, err := ParseTrainees(rows); err == nil {
		t.Fatal("expected an error for an unrecognized status, got nil")
	}
}

func TestRosterRoundTrip(t *testing.T) {
	original := []Trainee{
		{
			Trainee:          assignment.Trainee{ID: "s1", Name: "Youssef", Gender: assignment.Male},
			CodeforcesHandle: "youssef_cf",
			Status:           StatusActive,
			WarningCount:     1,
			TelegramChatID:   555,
		},
	}
	rows := RosterRows(original)
	parsed, err := ParseTrainees(rows)
	if err != nil {
		t.Fatalf("unexpected error round-tripping: %v", err)
	}
	if len(parsed) != 1 || parsed[0] != original[0] {
		t.Errorf("round trip mismatch: got %+v, want %+v", parsed, original)
	}
}
