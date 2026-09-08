// Package policy decides what should happen to a trainee given their
// current roster status and their Codeforces activity — separate from
// how that activity was fetched or how the decision gets carried out.
// Keeping this pure (no I/O) means every threshold and edge case can
// be unit-tested directly, without a live Codeforces handle or a real
// spreadsheet.
package policy

import (
	"fmt"

	"fcds-training-tool/internal/Codeforces"
	"fcds-training-tool/internal/roster"
)

// Config holds the tunable thresholds behind every decision. Start
// from DefaultConfig and adjust to match how the community actually
// wants to run the program — these numbers are a starting point, not
// a claim about what's "correct".
type Config struct {
	// MinSolvedForActive is the fewest problems an active trainee must
	// solve within the tracked period to avoid a warning.
	MinSolvedForActive int
	// MaxWarningsBeforeFilter is how many consecutive low-activity
	// periods an active trainee can accumulate before being filtered.
	MaxWarningsBeforeFilter int
	// MinSolvedForPromotion is the fewest problems a waitlisted
	// trainee must solve within the tracked period to be flagged as a
	// promotion candidate.
	MinSolvedForPromotion int
}

func DefaultConfig() Config {
	return Config{
		MinSolvedForActive:      1,
		MaxWarningsBeforeFilter: 2,
		MinSolvedForPromotion:   3,
	}
}

type Action string

const (
	ActionNone    Action = "none"
	ActionWarn    Action = "warn"
	ActionFilter  Action = "filter"
	ActionPromote Action = "promote"
)

// Decision is the outcome of evaluating one trainee.
type Decision struct {
	TraineeID string
	Action    Action
	Reason    string
}

// Evaluate decides what should happen to a single trainee given their
// current roster state and their Codeforces activity summary. It does
// not mutate anything or send any message — callers apply the
// decision (update roster status, queue a notification) themselves.
func Evaluate(cfg Config, t roster.Trainee, activity codeforces.ActivitySummary) Decision {
	switch t.Status {
	case roster.StatusActive:
		if activity.SolvedInPeriod >= cfg.MinSolvedForActive {
			return Decision{TraineeID: t.ID, Action: ActionNone,
				Reason: fmt.Sprintf("solved %d problems, meets the threshold of %d", activity.SolvedInPeriod, cfg.MinSolvedForActive)}
		}
		if t.WarningCount+1 >= cfg.MaxWarningsBeforeFilter {
			return Decision{TraineeID: t.ID, Action: ActionFilter,
				Reason: fmt.Sprintf("inactive for %d consecutive periods (limit %d)", t.WarningCount+1, cfg.MaxWarningsBeforeFilter)}
		}
		return Decision{TraineeID: t.ID, Action: ActionWarn,
			Reason: fmt.Sprintf("solved %d problems, below the threshold of %d", activity.SolvedInPeriod, cfg.MinSolvedForActive)}

	case roster.StatusWaitlisted:
		if activity.SolvedInPeriod >= cfg.MinSolvedForPromotion {
			return Decision{TraineeID: t.ID, Action: ActionPromote,
				Reason: fmt.Sprintf("solved %d problems while waitlisted (threshold %d)", activity.SolvedInPeriod, cfg.MinSolvedForPromotion)}
		}
		return Decision{TraineeID: t.ID, Action: ActionNone,
			Reason: fmt.Sprintf("solved %d problems, below the promotion threshold of %d", activity.SolvedInPeriod, cfg.MinSolvedForPromotion)}

	default: // StatusFiltered, or anything else: nothing left to decide
		return Decision{TraineeID: t.ID, Action: ActionNone, Reason: "no action for a filtered trainee"}
	}
}
