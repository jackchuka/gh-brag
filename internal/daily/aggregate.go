package daily

import (
	"sort"
	"strings"

	"github.com/jackchuka/gh-brag/internal/data"
)

// Aggregate combines PRs and reviews into a DailyReport
// PRs are grouped by linked issue; PRs without issues are standalone
func Aggregate(dateRange *DateRange, prs []PRWithIssues, reviews []ReviewedPR) *DailyReport {
	report := &DailyReport{
		DateLabel:     dateRange.Label,
		RangeStart:    dateRange.Start,
		RangeEnd:      dateRange.End,
		IssueGroups:   make([]IssueGroup, 0),
		StandalonePRs: make([]data.Event, 0),
		ExtraReviews:  make([]ExtraReview, 0),
	}

	// Group PRs by linked issue
	// Key: issue URL (unique identifier)
	issueMap := make(map[string]*IssueGroup)
	var issueOrder []string // preserve order of first appearance

	for _, pr := range prs {
		if len(pr.LinkedIssues) == 0 {
			// No linked issues - standalone PR
			report.StandalonePRs = append(report.StandalonePRs, pr.Event)
		} else {
			// Add PR to each linked issue group
			for _, issue := range pr.LinkedIssues {
				if group, exists := issueMap[issue.URL]; exists {
					group.PRs = append(group.PRs, pr.Event)
				} else {
					issueMap[issue.URL] = &IssueGroup{
						Issue: issue,
						PRs:   []data.Event{pr.Event},
					}
					issueOrder = append(issueOrder, issue.URL)
				}
			}
		}
	}

	// Build issue groups in order of first appearance
	for _, url := range issueOrder {
		group := issueMap[url]
		// Sort PRs within group by activity time descending
		sort.Slice(group.PRs, func(i, j int) bool {
			return ActivityTime(group.PRs[i]).After(ActivityTime(group.PRs[j]))
		})
		report.IssueGroups = append(report.IssueGroups, *group)
	}

	// Sort issue groups by most recent PR activity descending
	sort.Slice(report.IssueGroups, func(i, j int) bool {
		iTime := ActivityTime(report.IssueGroups[i].PRs[0])
		jTime := ActivityTime(report.IssueGroups[j].PRs[0])
		return iTime.After(jTime)
	})

	// Sort standalone PRs by activity time descending
	sort.Slice(report.StandalonePRs, func(i, j int) bool {
		return ActivityTime(report.StandalonePRs[i]).After(ActivityTime(report.StandalonePRs[j]))
	})

	// Convert ReviewedPR to ExtraReview (grouped by PR)
	for _, reviewed := range reviews {
		parts := strings.Split(reviewed.Event.Repo, "/")
		owner := ""
		repo := reviewed.Event.Repo
		if len(parts) == 2 {
			owner = parts[0]
			repo = parts[1]
		}

		// Sort reviews by submittedAt ascending within each PR
		sortedReviews := make([]ReviewInfo, len(reviewed.Reviews))
		copy(sortedReviews, reviewed.Reviews)
		sort.Slice(sortedReviews, func(i, j int) bool {
			return sortedReviews[i].SubmittedAt.Before(sortedReviews[j].SubmittedAt)
		})

		report.ExtraReviews = append(report.ExtraReviews, ExtraReview{
			Owner:    owner,
			Repo:     repo,
			PRNumber: reviewed.Event.Number,
			PRTitle:  reviewed.Event.Title,
			PRURL:    reviewed.Event.URL,
			Reviews:  sortedReviews,
		})
	}

	// Sort extra reviews by most recent review submittedAt descending
	sort.Slice(report.ExtraReviews, func(i, j int) bool {
		iLast := report.ExtraReviews[i].Reviews[len(report.ExtraReviews[i].Reviews)-1]
		jLast := report.ExtraReviews[j].Reviews[len(report.ExtraReviews[j].Reviews)-1]
		return iLast.SubmittedAt.After(jLast.SubmittedAt)
	})

	return report
}
