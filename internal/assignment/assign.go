package assignment

import "sort"

// Assign computes a trainer/trainee assignment that:
//   - gives every trainer as close to an equal share of trainees as
//     possible (quotas differ by at most one trainee), and
//   - never assigns a male trainee to a female trainer (female
//     trainers only take female trainees; male trainers can take
//     trainees of either gender).
//
// It models the problem as a max-flow instance:
//
//	source -> trainee (cap 1) -> eligible trainer (cap 1) -> sink (cap = trainer's quota)
//
// and reads the assignment off the saturated trainee->trainer edges.
// If the constraints make a full assignment impossible (e.g. more
// male trainees than male-trainer capacity), the trainees that could
// not be placed are returned in Assignment.Unassigned instead of
// being silently dropped or force-matched incorrectly.
func Assign(trainers []Trainer, trainees []Trainee) Assignment {
	n := len(trainers)
	m := len(trainees)

	source := 0
	traineeNode := func(i int) int { return 1 + i }
	trainerNode := func(j int) int { return 1 + m + j }
	sink := 1 + m + n

	g := newFlowGraph(sink + 1)

	for i := range trainees {
		g.addEdge(source, traineeNode(i), 1)
	}

	quotas := fairQuotas(n, m)
	for j := range trainers {
		g.addEdge(trainerNode(j), sink, quotas[j])
	}

	for i, trainee := range trainees {
		for j, trainer := range trainers {
			if trainer.Gender == Female && trainee.Gender == Male {
				continue // female trainers only take female trainees
			}
			g.addEdge(traineeNode(i), trainerNode(j), 1)
		}
	}

	g.maxFlow(source, sink)

	result := Assignment{TraineeToTrainer: make(map[string]string)}
	for i, trainee := range trainees {
		placed := false
		for _, ei := range g.edges[traineeNode(i)] {
			e := g.all[ei]
			if e.flow > 0 && e.to >= trainerNode(0) && e.to < sink {
				j := e.to - trainerNode(0)
				result.TraineeToTrainer[trainee.ID] = trainers[j].ID
				placed = true
				break
			}
		}
		if !placed {
			result.Unassigned = append(result.Unassigned, trainee.ID)
		}
	}
	sort.Strings(result.Unassigned)
	return result
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
