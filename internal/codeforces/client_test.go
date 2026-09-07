package codeforces

import (
	"testing"
	"time"
)

func TestSummarize(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	submissions := []Submission{
		{CreationTimeSeconds: now.Add(-2 * time.Hour).Unix(), Verdict: "OK", Problem: Problem{ContestID: 1, Index: "A"}},
		{CreationTimeSeconds: now.Add(-90 * time.Minute).Unix(), Verdict: "WRONG_ANSWER", Problem: Problem{ContestID: 1, Index: "B"}},
		{CreationTimeSeconds: now.Add(-60 * time.Minute).Unix(), Verdict: "OK", Problem: Problem{ContestID: 1, Index: "A"}},
		{CreationTimeSeconds: now.Add(-10 * 24 * time.Hour).Unix(), Verdict: "OK", Problem: Problem{ContestID: 2, Index: "A"}},
	}

	got := Summarize("tourist", submissions, now, 7*24*time.Hour)
	if got.TotalSubmissions != 4 {
		t.Fatalf("TotalSubmissions = %d, want 4", got.TotalSubmissions)
	}
	if got.AcceptedSubmissions != 3 {
		t.Fatalf("AcceptedSubmissions = %d, want 3", got.AcceptedSubmissions)
	}
	if got.UniqueSolved != 2 {
		t.Fatalf("UniqueSolved = %d, want 2", got.UniqueSolved)
	}
	if got.SolvedInPeriod != 1 {
		t.Fatalf("SolvedInPeriod = %d, want 1", got.SolvedInPeriod)
	}
}
