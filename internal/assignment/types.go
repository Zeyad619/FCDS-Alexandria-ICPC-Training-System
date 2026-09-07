package assignment

// Gender is used only for the trainer/trainee matching constraint:
// a female trainer should only be assigned female trainees. Male
// trainers may take trainees of either gender.
type Gender string

const (
	Female Gender = "female"
	Male   Gender = "male"
)

// Trainer is a CP trainer available to take on trainees.
type Trainer struct {
	ID     string
	Name   string
	Gender Gender
}

// Trainee is a CP trainee waiting to be assigned to a trainer.
type Trainee struct {
	ID     string
	Name   string
	Gender Gender
}

// Assignment is the result of running Assign: which trainer each
// trainee was matched to, plus any trainees that could not be placed
// at all given the current trainers and constraints.
type Assignment struct {
	TraineeToTrainer map[string]string
	Unassigned       []string
}
