package daily

import (
	"time"

	"github.com/jackchuka/gh-brag/internal/data"
)

// LinkedIssue represents an issue linked to a PR via closingIssuesReferences
type LinkedIssue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

// ReviewInfo represents a review you submitted on a PR
type ReviewInfo struct {
	State       string    `json:"state"` // APPROVED | CHANGES_REQUESTED | COMMENTED
	SubmittedAt time.Time `json:"submittedAt"`
	URL         string    `json:"url,omitempty"`
}

// IssueGroup represents an issue with all PRs that link to it
type IssueGroup struct {
	Issue LinkedIssue  `json:"issue"`
	PRs   []data.Event `json:"prs"`
}

// ExtraReview represents reviews you submitted on someone else's PR (grouped by PR)
type ExtraReview struct {
	Owner    string       `json:"owner"`
	Repo     string       `json:"repo"`
	PRNumber int          `json:"prNumber"`
	PRTitle  string       `json:"prTitle"`
	PRURL    string       `json:"prUrl"`
	Reviews  []ReviewInfo `json:"reviews"`
}

// DailyReport is the top-level output structure
type DailyReport struct {
	Summary       string        `json:"summary,omitempty"` // LLM-generated summary
	DateLabel     string        `json:"dateLabel"`         // YYYY-MM-DD or YYYY-MM-DD..YYYY-MM-DD
	RangeStart    time.Time     `json:"rangeStart"`
	RangeEnd      time.Time     `json:"rangeEnd"`
	IssueGroups   []IssueGroup  `json:"issueGroups"`   // PRs grouped by linked issue
	StandalonePRs []data.Event  `json:"standalonePrs"` // PRs without linked issues
	ExtraReviews  []ExtraReview `json:"extraReviews"`
}

// PRWithIssues wraps a PR event with its linked issues (intermediate fetch type)
type PRWithIssues struct {
	Event        data.Event
	LinkedIssues []LinkedIssue
}

// ReviewedPR wraps a reviewed PR with your review details (intermediate fetch type)
type ReviewedPR struct {
	Event   data.Event
	Reviews []ReviewInfo
}

// activityTime returns the timestamp used for filtering and sorting
func ActivityTime(e data.Event) time.Time {
	// For merged PRs, ClosedAt equals merge time
	if e.Action == data.EventActionMerged && !e.Timestamps.ClosedAt.IsZero() {
		return e.Timestamps.ClosedAt
	}
	if !e.Timestamps.UpdatedAt.IsZero() {
		return e.Timestamps.UpdatedAt
	}
	return e.Timestamps.CreatedAt
}
