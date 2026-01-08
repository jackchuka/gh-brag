package daily

import (
	"fmt"
	"time"

	"github.com/jackchuka/gh-brag/internal/data"
	"github.com/jackchuka/gh-brag/internal/github"
)

// FetchAuthoredPRs fetches PRs authored by the current user with linked issues
// Uses open-ended query (updated >= start) and filters locally by activity time
func FetchAuthoredPRs(dateRange *DateRange, includeLinkedIssues bool, orgs []string) ([]PRWithIssues, error) {
	// Open-ended query: fetch all PRs updated since start, filter end locally
	baseQuery := fmt.Sprintf("author:@me is:pr updated:%s", dateRange.FormatStartForGitHub())

	// Build list of queries (one per org, or single query if no orgs)
	queries := buildQueries(baseQuery, orgs)

	var results []PRWithIssues
	seen := make(map[string]bool)
	fetchedAt := time.Now()

	for _, query := range queries {
		nodes, err := github.RunSearch(query, github.QueryWithLinkedIssues)
		if err != nil {
			return nil, err
		}

		for _, n := range nodes {
			if n.Typename != "PullRequest" {
				continue
			}

			// Dedupe by URL
			if seen[n.URL] {
				continue
			}
			seen[n.URL] = true

			// Determine action based on state
			action := data.EventActionAuthored
			if !n.MergedAt.IsZero() {
				action = data.EventActionMerged
			}

			evt := data.Event{
				ID:     fmt.Sprintf("pr:%s:%s", n.URL, action),
				Action: action,
				Kind:   "pr",
				URL:    n.URL,
				Repo:   n.Repository.NameWithOwner,
				Number: n.Number,
				Title:  n.Title,
				Body:   n.Body,
				Author: n.Author.Login,
				Timestamps: data.Timestamps{
					CreatedAt: n.CreatedAt,
					UpdatedAt: n.UpdatedAt,
					ClosedAt:  n.ClosedAt,
				},
				Source: data.Source{
					Tool:      "gh api graphql",
					Query:     query,
					FetchedAt: fetchedAt,
				},
			}

			// Filter by activity time: must be within date range [Start, End)
			activityTime := ActivityTime(evt)
			if activityTime.Before(dateRange.Start) || !activityTime.Before(dateRange.End) {
				continue
			}

			// Extract labels
			labels := make([]string, 0, len(n.Labels.Nodes))
			for _, l := range n.Labels.Nodes {
				labels = append(labels, l.Name)
			}
			evt.Labels = labels

			// Extract linked issues
			var linkedIssues []LinkedIssue
			if includeLinkedIssues {
				for _, issue := range n.ClosingIssuesReferences.Nodes {
					linkedIssues = append(linkedIssues, LinkedIssue{
						Number: issue.Number,
						Title:  issue.Title,
						URL:    issue.URL,
					})
				}
			}

			results = append(results, PRWithIssues{
				Event:        evt,
				LinkedIssues: linkedIssues,
			})
		}
	}

	return results, nil
}

// FetchReviewedPRs fetches PRs reviewed by the current user with review details
// Uses open-ended query (updated >= start) and filters locally by review submittedAt
func FetchReviewedPRs(dateRange *DateRange, currentUser string, orgs []string) ([]ReviewedPR, error) {
	// Open-ended query: fetch all reviewed PRs since start, filter end locally by submittedAt
	baseQuery := fmt.Sprintf("is:pr reviewed-by:@me updated:%s -author:@me", dateRange.FormatStartForGitHub())

	// Build list of queries (one per org, or single query if no orgs)
	queries := buildQueries(baseQuery, orgs)

	var results []ReviewedPR
	seen := make(map[string]bool)
	fetchedAt := time.Now()

	for _, query := range queries {
		nodes, err := github.RunSearch(query, github.QueryWithReviews)
		if err != nil {
			return nil, err
		}

		for _, n := range nodes {
			if n.Typename != "PullRequest" {
				continue
			}

			// Dedupe by URL
			if seen[n.URL] {
				continue
			}
			seen[n.URL] = true

			evt := data.Event{
				ID:     fmt.Sprintf("pr:%s:reviewed", n.URL),
				Action: data.EventActionReviewed,
				Kind:   "pr",
				URL:    n.URL,
				Repo:   n.Repository.NameWithOwner,
				Number: n.Number,
				Title:  n.Title,
				Author: n.Author.Login,
				Timestamps: data.Timestamps{
					CreatedAt: n.CreatedAt,
					UpdatedAt: n.UpdatedAt,
					ClosedAt:  n.ClosedAt,
				},
				Source: data.Source{
					Tool:      "gh api graphql",
					Query:     query,
					FetchedAt: fetchedAt,
				},
			}

			// Extract reviews by the current user within the date range
			var reviews []ReviewInfo
			for _, r := range n.Reviews.Nodes {
				// Filter to only include reviews by the current user
				if r.Author.Login == currentUser || currentUser == "" {
					// Skip PENDING reviews
					if r.State == "PENDING" {
						continue
					}
					// Filter by date range
					if r.SubmittedAt.Before(dateRange.Start) || !r.SubmittedAt.Before(dateRange.End) {
						continue
					}
					reviews = append(reviews, ReviewInfo{
						State:       r.State,
						SubmittedAt: r.SubmittedAt,
						URL:         r.URL,
					})
				}
			}

			// Only include if there are reviews
			if len(reviews) > 0 {
				results = append(results, ReviewedPR{
					Event:   evt,
					Reviews: reviews,
				})
			}
		}
	}

	return results, nil
}

// buildQueries returns a list of queries to execute.
// If orgs is empty, returns the base query as-is.
// Otherwise, returns one query per org with the org filter appended.
func buildQueries(baseQuery string, orgs []string) []string {
	if len(orgs) == 0 {
		return []string{baseQuery}
	}
	queries := make([]string, len(orgs))
	for i, org := range orgs {
		queries[i] = baseQuery + fmt.Sprintf(" org:%s", org)
	}
	return queries
}
