// Package roster tracks the extra state the assignment engine doesn't
// need: each trainee's Codeforces handle, Telegram chat ID, whether
// they're active or on the waiting list, and how many warnings they've
// accumulated. The matching engine in internal/assignment only ever
// sees plain assignment.Trainee values with no notion of status.
package roster

import "fcds-training-tool/internal/assignment"

// Status is where a trainee currently stands in the program.
type Status string

const (
	StatusActive     Status = "active"
	StatusWaitlisted Status = "waitlisted"
	StatusFiltered   Status = "filtered"
)

// Trainee is the roster's view of a trainee: identity (via the embedded
// assignment.Trainee) plus the extra state needed to decide warnings,
// filtering, and waiting-list promotion.
type Trainee struct {
	assignment.Trainee
	CodeforcesHandle string
	TelegramChatID   int64
	Status           Status
	WarningCount     int
}

// ActiveAssignmentInput returns the plain assignment.Trainee values for
// everyone currently active, in the shape the matching engine expects.
// Waitlisted and filtered trainees are excluded — they don't compete
// for trainer slots.
func ActiveAssignmentInput(trainees []Trainee) []assignment.Trainee {
	var active []assignment.Trainee
	for _, t := range trainees {
		if t.Status == StatusActive {
			active = append(active, t.Trainee)
		}
	}
	return active
}
