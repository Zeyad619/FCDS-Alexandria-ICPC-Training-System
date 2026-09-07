// Command assign runs the trainer/trainee assignment engine against
// sample data and prints the result. This is a temporary demo entry
// point: the sample data below will be replaced by data read from a
// live Google Sheet once that integration is built (see README).
package main

import (
	"fmt"
	"sort"

	"fcds-training-tool/internal/assignment"
)

func sampleData() ([]assignment.Trainer, []assignment.Trainee) {
	trainers := []assignment.Trainer{
		{ID: "t1", Name: "Ahmed", Gender: assignment.Male},
		{ID: "t2", Name: "Sara", Gender: assignment.Female},
		{ID: "t3", Name: "Omar", Gender: assignment.Male},
	}
	trainees := []assignment.Trainee{
		{ID: "s1", Name: "Youssef", Gender: assignment.Male},
		{ID: "s2", Name: "Mona", Gender: assignment.Female},
		{ID: "s3", Name: "Karim", Gender: assignment.Male},
		{ID: "s4", Name: "Laila", Gender: assignment.Female},
		{ID: "s5", Name: "Hassan", Gender: assignment.Male},
		{ID: "s6", Name: "Nour", Gender: assignment.Female},
		{ID: "s7", Name: "Mahmoud", Gender: assignment.Male},
	}
	return trainers, trainees
}

func main() {
	trainers, trainees := sampleData()
	result := assignment.Assign(trainers, trainees)

	byTrainer := make(map[string][]string)
	for traineeID, trainerID := range result.TraineeToTrainer {
		byTrainer[trainerID] = append(byTrainer[trainerID], traineeID)
	}

	for _, tr := range trainers {
		names := byTrainer[tr.ID]
		sort.Strings(names)
		fmt.Printf("%s (%s trainer): %v\n", tr.Name, tr.Gender, names)
	}
	if len(result.Unassigned) > 0 {
		fmt.Printf("Unassigned: %v\n", result.Unassigned)
	}
