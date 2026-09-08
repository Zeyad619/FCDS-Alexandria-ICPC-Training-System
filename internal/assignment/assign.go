package assignment

import "sort"

// Assign computes a trainer/trainee assignment with a two-phase,
// two-pointer approach — no flow network, just two linear passes over
// gender-partitioned slices:
//
//  1. Partition trainers and trainees by gender (female first). A
//     stable partition on a two-valued key is exactly a stable sort by
//     that key, done in O(N) instead of O(N log N).
//  2. Phase 1: walk a trainer-pointer and a trainee-pointer together
//     over the female trainers and female trainees only, filling each
//     female trainer's fair-share quota from the female-trainee pool
//     (female trainers can never take a male trainee).
//  3. Phase 2: whatever's left over — any female trainees a female
//     trainer's quota couldn't reach, plus every male trainee — is
//     walked the same way against the male trainers. Male-trainer
//     quotas are recomputed against exactly what's left in phase 2,
//     so capacity a female trainer couldn't use (because there
//     weren't enough female trainees) is picked up by the male
//     trainers instead of sitting idle.
//
// Because male trainers accept trainees of either gender, giving
// female trainees to female trainers first never costs the male
// trainers a placement they'd otherwise have made — it can only free
// up room. That's what makes the two-phase split correct here without
// needing a general matching/flow algorithm: the eligibility
// structure is "one group takes anyone, the other only takes a
// subset", not an arbitrary bipartite graph.
//
// Anyone still unplaced after phase 2 (only possible when there are
// no male trainers at all to absorb male trainees or leftover female
// trainees) comes back in Assignment.Unassigned instead of being
// silently dropped.
func Assign(trainers []Trainer, trainees []Trainee) Assignment {
	femaleTrainers, maleTrainers := splitTrainersByGender(trainers)
	femaleTrainees, maleTrainees := splitTraineesByGender(trainees)

	result := Assignment{TraineeToTrainer: make(map[string]string)}

	// Phase 1: female trainers <- female trainees only.
	aspirationalQuotas := fairQuotas(len(trainers), len(trainees))
	femaleQuotas := aspirationalQuotas[:len(femaleTrainers)]

	traineePtr := 0
	for trainerIdx, quota := range femaleQuotas {
		for taken := 0; taken < quota && traineePtr < len(femaleTrainees); taken++ {
			result.TraineeToTrainer[femaleTrainees[traineePtr].ID] = femaleTrainers[trainerIdx].ID
			traineePtr++
		}
	}
	placedByFemaleTrainers := traineePtr

	// Phase 2: leftover female trainees + all male trainees <- male trainers,
	// with male-trainer quotas recomputed against what's actually left.
	remaining := make([]Trainee, 0, len(femaleTrainees)-placedByFemaleTrainers+len(maleTrainees))
	remaining = append(remaining, femaleTrainees[placedByFemaleTrainers:]...)
	remaining = append(remaining, maleTrainees...)

	maleQuotas := fairQuotas(len(maleTrainers), len(remaining))
	remainingPtr := 0
	for trainerIdx, quota := range maleQuotas {
		for taken := 0; taken < quota && remainingPtr < len(remaining); taken++ {
			result.TraineeToTrainer[remaining[remainingPtr].ID] = maleTrainers[trainerIdx].ID
			remainingPtr++
		}
	}

	for _, t := range remaining[remainingPtr:] {
		result.Unassigned = append(result.Unassigned, t.ID)
	}
	sort.Strings(result.Unassigned)
	return result
}

// splitTrainersByGender partitions trainers into (female, male),
// preserving each group's relative order — a stable partition, i.e. a
// stable sort by the two-valued "is female" key.
func splitTrainersByGender(trainers []Trainer) (female, male []Trainer) {
	for _, t := range trainers {
		if t.Gender == Female {
			female = append(female, t)
		} else {
			male = append(male, t)
		}
	}
	return female, male
}

func splitTraineesByGender(trainees []Trainee) (female, male []Trainee) {
	for _, t := range trainees {
		if t.Gender == Female {
			female = append(female, t)
		} else {
			male = append(male, t)
		}
	}
	return female, male
}

// fairQuotas splits m items across n buckets as evenly as possible:
// every bucket gets base = m/n, and the first (m%n) buckets get one
// extra so the total adds up to exactly m.
func fairQuotas(n, m int) []int {
	if n == 0 {
		return nil
	}
	base := m / n
	remainder := m % n
	quotas := make([]int, n)
	for i := range quotas {
		quotas[i] = base
		if i < remainder {
			quotas[i]++
		}
	}
	return quotas
}
