package assignment

import "testing"

func TestAssign_EvenSplitNoGenderConstraint(t *testing.T) {
	trainers := []Trainer{
		{ID: "t1", Gender: Male},
		{ID: "t2", Gender: Male},
	}
	trainees := []Trainee{
		{ID: "s1", Gender: Male},
		{ID: "s2", Gender: Male},
		{ID: "s3", Gender: Male},
		{ID: "s4", Gender: Male},
	}
	result := Assign(trainers, trainees)

	if len(result.Unassigned) != 0 {
		t.Fatalf("expected everyone assigned, got unassigned: %v", result.Unassigned)
	}
	counts := map[string]int{}
	for _, trainerID := range result.TraineeToTrainer {
		counts[trainerID]++
	}
	for _, tr := range trainers {
		if counts[tr.ID] != 2 {
			t.Errorf("trainer %s: expected 2 trainees, got %d", tr.ID, counts[tr.ID])
		}
	}
}

func TestAssign_FemaleTrainerOnlyGetsFemaleTrainees(t *testing.T) {
	trainers := []Trainer{
		{ID: "female-trainer", Gender: Female},
		{ID: "male-trainer", Gender: Male},
	}
	trainees := []Trainee{
		{ID: "f1", Gender: Female},
		{ID: "f2", Gender: Female},
		{ID: "m1", Gender: Male},
		{ID: "m2", Gender: Male},
	}
	result := Assign(trainers, trainees)

	if len(result.Unassigned) != 0 {
		t.Fatalf("expected everyone assigned, got unassigned: %v", result.Unassigned)
	}
	if result.TraineeToTrainer["f1"] != "female-trainer" || result.TraineeToTrainer["f2"] != "female-trainer" {
		t.Errorf("expected both female trainees with female-trainer, got f1=%s f2=%s",
			result.TraineeToTrainer["f1"], result.TraineeToTrainer["f2"])
	}
	if result.TraineeToTrainer["m1"] != "male-trainer" || result.TraineeToTrainer["m2"] != "male-trainer" {
		t.Errorf("expected both male trainees with male-trainer, got m1=%s m2=%s",
			result.TraineeToTrainer["m1"], result.TraineeToTrainer["m2"])
	}
}

// This is the case a naive single-pass two-pointer (sort everyone by
// gender, then hand out `quota` trainees per trainer in order) gets
// wrong: with only 1 female trainee against a female trainer whose
// aspirational quota is 2, a naive walk would keep taking from the
// trainee list and hand her a male trainee next. The two-phase
// version must never do that.
func TestAssign_NeverGivesFemaleTrainerAMaleTrainee(t *testing.T) {
	trainers := []Trainer{
		{ID: "female-trainer", Gender: Female},
		{ID: "male-trainer-1", Gender: Male},
		{ID: "male-trainer-2", Gender: Male},
	}
	trainees := []Trainee{
		{ID: "f1", Gender: Female},
		{ID: "m1", Gender: Male},
		{ID: "m2", Gender: Male},
		{ID: "m3", Gender: Male},
	}
	result := Assign(trainers, trainees)

	if len(result.Unassigned) != 0 {
		t.Fatalf("expected everyone assigned (male trainers can absorb the shortfall), got unassigned: %v", result.Unassigned)
	}
	if result.TraineeToTrainer["f1"] != "female-trainer" {
		t.Errorf("expected f1 with female-trainer, got %s", result.TraineeToTrainer["f1"])
	}
	for _, id := range []string{"m1", "m2", "m3"} {
		trainerID := result.TraineeToTrainer[id]
		if trainerID == "female-trainer" {
			t.Errorf("trainee %s (male) was assigned to the female trainer — constraint violated", id)
		}
	}
}

// Unused female-trainer capacity (she wanted 2, only 1 female trainee
// existed) must be picked up by the male trainers rather than wasted:
// this scenario is fully placeable, unlike the old fixed-quota design
// which would have left one trainee stranded.
func TestAssign_ReallocatesUnusedFemaleTrainerCapacity(t *testing.T) {
	trainers := []Trainer{
		{ID: "female-trainer", Gender: Female}, // aspirational quota 2
		{ID: "male-trainer", Gender: Male},     // aspirational quota 1
	}
	trainees := []Trainee{
		{ID: "f1", Gender: Female},
		{ID: "m1", Gender: Male},
		{ID: "m2", Gender: Male},
	}
	result := Assign(trainers, trainees)

	if len(result.Unassigned) != 0 {
		t.Fatalf("expected everyone placed via reallocation, got unassigned: %v", result.Unassigned)
	}
	if result.TraineeToTrainer["f1"] != "female-trainer" {
		t.Errorf("expected f1 with female-trainer, got %s", result.TraineeToTrainer["f1"])
	}
	if result.TraineeToTrainer["m1"] != "male-trainer" || result.TraineeToTrainer["m2"] != "male-trainer" {
		t.Errorf("expected both male trainees with male-trainer, got m1=%s m2=%s",
			result.TraineeToTrainer["m1"], result.TraineeToTrainer["m2"])
	}
}

// The only way someone can genuinely be left unplaced is when there
// are no male trainers at all to absorb male trainees (or leftover
// female trainees) — female trainers simply cannot take them.
func TestAssign_ReportsUnassignedWhenNoMaleTrainerExists(t *testing.T) {
	trainers := []Trainer{
		{ID: "female-trainer", Gender: Female},
	}
	trainees := []Trainee{
		{ID: "f1", Gender: Female},
		{ID: "f2", Gender: Female},
		{ID: "m1", Gender: Male},
	}
	result := Assign(trainers, trainees)

	if len(result.Unassigned) != 1 || result.Unassigned[0] != "m1" {
		t.Fatalf("expected exactly m1 unassigned, got %v", result.Unassigned)
	}
	if result.TraineeToTrainer["f1"] != "female-trainer" || result.TraineeToTrainer["f2"] != "female-trainer" {
		t.Errorf("expected both female trainees with the female trainer, got f1=%s f2=%s",
			result.TraineeToTrainer["f1"], result.TraineeToTrainer["f2"])
	}
}

func TestFairQuotas(t *testing.T) {
	quotas := fairQuotas(3, 10)
	total := 0
	for _, q := range quotas {
		total += q
		if q < 3 || q > 4 {
			t.Errorf("quota %d out of expected [3,4] range", q)
		}
	}
	if total != 10 {
		t.Errorf("quotas sum to %d, expected 10", total)
	}
}
