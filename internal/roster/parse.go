package roster

import (
	"fmt"
	"strconv"
	"strings"

	"fcds-training-tool/internal/assignment"
)

// header builds a case-insensitive column-name -> index lookup from a
// sheet's first row, so parsing doesn't depend on column order.
func header(row []string) map[string]int {
	idx := make(map[string]int, len(row))
	for i, name := range row {
		idx[strings.ToLower(strings.TrimSpace(name))] = i
	}
	return idx
}

func cell(row []string, idx map[string]int, column string) string {
	i, ok := idx[column]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func parseGender(value string) (assignment.Gender, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "female", "f":
		return assignment.Female, nil
	case "male", "m":
		return assignment.Male, nil
	default:
		return "", fmt.Errorf("unrecognized gender %q (expected male/female)", value)
	}
}

func parseStatus(value string) (Status, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "active":
		return StatusActive, nil
	case "waitlisted", "waiting", "waitlist":
		return StatusWaitlisted, nil
	case "filtered":
		return StatusFiltered, nil
	default:
		return "", fmt.Errorf("unrecognized status %q (expected active/waitlisted/filtered)", value)
	}
}

// ParseTrainers reads rows where the first row is a header containing
// (case-insensitively, in any order): ID, Name, Gender.
func ParseTrainers(rows [][]string) ([]assignment.Trainer, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	idx := header(rows[0])
	trainers := make([]assignment.Trainer, 0, len(rows)-1)
	for r, row := range rows[1:] {
		id := cell(row, idx, "id")
		if id == "" {
			continue // skip blank trailing rows
		}
		gender, err := parseGender(cell(row, idx, "gender"))
		if err != nil {
			return nil, fmt.Errorf("trainer row %d (%s): %w", r+2, id, err)
		}
		trainers = append(trainers, assignment.Trainer{
			ID:     id,
			Name:   cell(row, idx, "name"),
			Gender: gender,
		})
	}
	return trainers, nil
}

// ParseTrainees reads rows where the first row is a header containing
// (case-insensitively, in any order): ID, Name, Gender, CodeforcesHandle,
// Status, TelegramChatID. TelegramChatID may be blank (some trainees may
// not have linked Telegram yet); everything else is required.
func ParseTrainees(rows [][]string) ([]Trainee, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	idx := header(rows[0])
	trainees := make([]Trainee, 0, len(rows)-1)
	for r, row := range rows[1:] {
		id := cell(row, idx, "id")
		if id == "" {
			continue
		}
		gender, err := parseGender(cell(row, idx, "gender"))
		if err != nil {
			return nil, fmt.Errorf("trainee row %d (%s): %w", r+2, id, err)
		}
		status, err := parseStatus(cell(row, idx, "status"))
		if err != nil {
			return nil, fmt.Errorf("trainee row %d (%s): %w", r+2, id, err)
		}

		var chatID int64
		if raw := cell(row, idx, "telegramchatid"); raw != "" {
			chatID, err = strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("trainee row %d (%s): invalid TelegramChatID %q: %w", r+2, id, raw, err)
			}
		}

		var warnings int
		if raw := cell(row, idx, "warningcount"); raw != "" {
			warnings, err = strconv.Atoi(raw)
			if err != nil {
				return nil, fmt.Errorf("trainee row %d (%s): invalid WarningCount %q: %w", r+2, id, raw, err)
			}
		}

		trainees = append(trainees, Trainee{
			Trainee: assignment.Trainee{
				ID:     id,
				Name:   cell(row, idx, "name"),
				Gender: gender,
			},
			CodeforcesHandle: cell(row, idx, "codeforceshandle"),
			TelegramChatID:   chatID,
			Status:           status,
			WarningCount:     warnings,
		})
	}
	return trainees, nil
}

// RosterRows converts the current trainee roster back into sheet rows
// (header + data), ready to write back with a Sheets client.
func RosterRows(trainees []Trainee) [][]string {
	rows := [][]string{{"ID", "Name", "Gender", "CodeforcesHandle", "Status", "WarningCount", "TelegramChatID"}}
	for _, t := range trainees {
		rows = append(rows, []string{
			t.ID,
			t.Name,
			string(t.Gender),
			t.CodeforcesHandle,
			string(t.Status),
			strconv.Itoa(t.WarningCount),
			strconv.FormatInt(t.TelegramChatID, 10),
		})
	}
	return rows
}
