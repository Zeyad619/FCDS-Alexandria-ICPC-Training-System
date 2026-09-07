package codeforces

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultBaseURL = "https://codeforces.com/api"

// Client is a small client for the public Codeforces API.
// The API can be queried anonymously for public user data.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// Submission contains the fields needed by the activity tracker.
type Submission struct {
	ID                   int64 `json:"id"`
	ContestID            int64 `json:"contestId"`
	CreationTimeSeconds  int64 `json:"creationTimeSeconds"`
	Problem              Problem `json:"problem"`
	Verdict              string `json:"verdict"`
}

// Problem identifies a Codeforces problem within a contest.
type Problem struct {
	ContestID int64  `json:"contestId"`
	Index     string `json:"index"`
	Name      string `json:"name"`
}

type apiResponse[T any] struct {
	Status  string `json:"status"`
	Comment string `json:"comment"`
	Result  T      `json:"result"`
}

// ActivitySummary is a compact, data-driven view of a user's recent work.
type ActivitySummary struct {
	Handle             string
	TotalSubmissions   int
	AcceptedSubmissions int
	UniqueSolved       int
	SolvedInPeriod     int
	LastSubmissionAt   time.Time
}

// NewClient creates a Codeforces client using the default public API endpoint.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{BaseURL: defaultBaseURL, HTTPClient: httpClient}
}

// UserStatus fetches a user's latest submissions. Public user status is available
// without an API key. Keep count reasonably small to respect API limits.
func (c *Client) UserStatus(ctx context.Context, handle string, count int) ([]Submission, error) {
	if handle == "" {
		return nil, fmt.Errorf("codeforces handle is required")
	}
	if count <= 0 {
		count = 100
	}

	endpoint, err := url.Parse(c.BaseURL + "/user.status")
	if err != nil {
		return nil, fmt.Errorf("parse Codeforces endpoint: %w", err)
	}
	q := endpoint.Query()
	q.Set("handle", handle)
	q.Set("from", "1")
	q.Set("count", strconv.Itoa(count))
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create Codeforces request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Codeforces API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Codeforces API returned HTTP %s", resp.Status)
	}

	var payload apiResponse[[]Submission]
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode Codeforces response: %w", err)
	}
	if payload.Status != "OK" {
		return nil, fmt.Errorf("Codeforces API failed: %s", payload.Comment)
	}
	return payload.Result, nil
}

// Summarize turns raw submissions into metrics that the training system can use
// for warnings and waiting-list decisions.
func Summarize(handle string, submissions []Submission, now time.Time, period time.Duration) ActivitySummary {
	summary := ActivitySummary{Handle: handle}
	if len(submissions) == 0 {
		return summary
	}

	if now.IsZero() {
		now = time.Now()
	}
	cutoff := now.Add(-period)
	seenSolved := make(map[string]struct{})
	seenSolvedInPeriod := make(map[string]struct{})

	for _, submission := range submissions {
		summary.TotalSubmissions++
		created := time.Unix(submission.CreationTimeSeconds, 0)
		if summary.LastSubmissionAt.IsZero() || created.After(summary.LastSubmissionAt) {
			summary.LastSubmissionAt = created
		}
		if submission.Verdict != "OK" {
			continue
		}
		summary.AcceptedSubmissions++
		problemKey := fmt.Sprintf("%d:%s", submission.Problem.ContestID, submission.Problem.Index)
		if _, exists := seenSolved[problemKey]; !exists {
			seenSolved[problemKey] = struct{}{}
			if created.After(cutoff) {
				seenSolvedInPeriod[problemKey] = struct{}{}
			}
		}
	}

	summary.UniqueSolved = len(seenSolved)
	summary.SolvedInPeriod = len(seenSolvedInPeriod)
	return summary
}
