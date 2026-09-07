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

// When there are more male trainees than male-trainer capacity, some
// male trainees must be left unassigned rather than incorrectly
// routed to a female trainer.
func TestAssign_ReportsUnassignedWhenInfeasible(t *testing.T) {
	trainers := []Trainer{
		{ID: "female-trainer", Gender: Female}, // quota will be 2
		{ID: "male-trainer", Gender: Male},     // quota will be 1
	}
	trainees := []Trainee{
		{ID: "f1", Gender: Female},
		{ID: "m1", Gender: Male},
		{ID: "m2", Gender: Male},
	}
	result := Assign(trainers, trainees)

	if len(result.Unassigned) != 1 {
		t.Fatalf("expected exactly 1 unassigned trainee, got %v", result.Unassigned)
	}
	if result.TraineeToTrainer["f1"] != "female-trainer" {
		t.Errorf("expected f1 with female-trainer, got %s", result.TraineeToTrainer["f1"])
	}
	for _, id := range []string{"m1", "m2"} {
		if trainerID, ok := result.TraineeToTrainer[id]; ok && trainerID != "male-trainer" {
			t.Errorf("trainee %s assigned to %s, expected male-trainer or unassigned", id, trainerID)
		}
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
