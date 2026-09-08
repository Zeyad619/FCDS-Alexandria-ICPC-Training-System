package policy

import (
	"testing"

	"fcds-training-tool/internal/Codeforces"
	"fcds-training-tool/internal/assignment"
	"fcds-training-tool/internal/roster"
)

func activeTrainee(warnings int) roster.Trainee {
	return roster.Trainee{
		Trainee:      assignment.Trainee{ID: "s1", Name: "Youssef"},
		Status:       roster.StatusActive,
		WarningCount: warnings,
	}
}

func waitlistedTrainee() roster.Trainee {
	return roster.Trainee{
		Trainee: assignment.Trainee{ID: "s2", Name: "Mona"},
		Status:  roster.StatusWaitlisted,
	}
}

func TestEvaluate_ActiveMeetsThreshold_NoAction(t *testing.T) {
	cfg := DefaultConfig()
	d := Evaluate(cfg, activeTrainee(0), codeforces.ActivitySummary{SolvedInPeriod: 2})
	if d.Action != ActionNone {
		t.Errorf("got %s, want none", d.Action)
	}
}

func TestEvaluate_ActiveBelowThreshold_FirstWarning(t *testing.T) {
	cfg := DefaultConfig() // MaxWarningsBeforeFilter = 2
	d := Evaluate(cfg, activeTrainee(0), codeforces.ActivitySummary{SolvedInPeriod: 0})
	if d.Action != ActionWarn {
		t.Errorf("got %s, want warn", d.Action)
	}
}

func TestEvaluate_ActiveBelowThreshold_FilteredAfterRepeatedWarnings(t *testing.T) {
	cfg := DefaultConfig() // MaxWarningsBeforeFilter = 2
	// Already warned once (WarningCount=1); a second low-activity
	// period should now filter them rather than warn again.
	d := Evaluate(cfg, activeTrainee(1), codeforces.ActivitySummary{SolvedInPeriod: 0})
	if d.Action != ActionFilter {
		t.Errorf("got %s, want filter", d.Action)
	}
}

func TestEvaluate_Waitlisted_PromotedWhenActiveEnough(t *testing.T) {
	cfg := DefaultConfig() // MinSolvedForPromotion = 3
	d := Evaluate(cfg, waitlistedTrainee(), codeforces.ActivitySummary{SolvedInPeriod: 5})
	if d.Action != ActionPromote {
		t.Errorf("got %s, want promote", d.Action)
	}
}

func TestEvaluate_Waitlisted_NoActionWhenBelowPromotionThreshold(t *testing.T) {
	cfg := DefaultConfig()
	d := Evaluate(cfg, waitlistedTrainee(), codeforces.ActivitySummary{SolvedInPeriod: 1})
	if d.Action != ActionNone {
		t.Errorf("got %s, want none", d.Action)
	}
}

func TestEvaluate_Filtered_AlwaysNoAction(t *testing.T) {
	cfg := DefaultConfig()
	filtered := roster.Trainee{Trainee: assignment.Trainee{ID: "s3"}, Status: roster.StatusFiltered}
	d := Evaluate(cfg, filtered, codeforces.ActivitySummary{SolvedInPeriod: 100})
	if d.Action != ActionNone {
		t.Errorf("got %s, want none — a filtered trainee should never get warned/promoted", d.Action)
	}
}
