package roster

// AssignmentRows converts a trainee->trainer assignment map into a
// sheet-friendly table (header + rows). Deliberately independent of
// internal/sheets: that package needs the Google API client library
// to build, and this conversion is pure data shaping that shouldn't
// have to drag that dependency along, or be untestable without it.
func AssignmentRows(assignments map[string]string) [][]string {
	rows := [][]string{{"Trainee ID", "Trainer ID"}}
	for traineeID, trainerID := range assignments {
		rows = append(rows, []string{traineeID, trainerID})
	}
	return rows
}

// ParseAssignments reads an assignment table back (header + rows,
// columns "Trainee ID" / "Trainer ID" in any order) into a lookup map.
// An empty sheet (no previous run yet) parses to an empty map, not an
// error.
func ParseAssignments(rows [][]string) (map[string]string, error) {
	result := make(map[string]string)
	if len(rows) == 0 {
		return result, nil
	}
	idx := header(rows[0])
	for _, row := range rows[1:] {
		traineeID := cell(row, idx, "trainee id")
		if traineeID == "" {
			continue
		}
		result[traineeID] = cell(row, idx, "trainer id")
	}
	return result, nil
}
