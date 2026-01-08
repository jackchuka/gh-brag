package daily

import (
	"testing"
	"time"

	"github.com/jackchuka/gh-brag/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregate_StandalonePRs(t *testing.T) {
	now := time.Now()
	dateRange := &DateRange{
		Start: now.Add(-24 * time.Hour),
		End:   now,
		Label: "2026-01-07",
	}

	prs := []PRWithIssues{
		{
			Event: data.Event{
				Title: "PR without issues",
				URL:   "https://github.com/org/repo/pull/1",
				Timestamps: data.Timestamps{
					UpdatedAt: now.Add(-1 * time.Hour),
				},
			},
			LinkedIssues: nil,
		},
	}

	report := Aggregate(dateRange, prs, nil)

	assert.Equal(t, "2026-01-07", report.DateLabel)
	assert.Len(t, report.StandalonePRs, 1)
	assert.Empty(t, report.IssueGroups)
	assert.Empty(t, report.ExtraReviews)
}

func TestAggregate_IssueGroups(t *testing.T) {
	now := time.Now()
	dateRange := &DateRange{
		Start: now.Add(-24 * time.Hour),
		End:   now,
		Label: "2026-01-07",
	}

	issue := LinkedIssue{
		Number: 100,
		Title:  "Fix bug",
		URL:    "https://github.com/org/repo/issues/100",
	}

	prs := []PRWithIssues{
		{
			Event: data.Event{
				Title: "PR fixing bug",
				URL:   "https://github.com/org/repo/pull/1",
				Timestamps: data.Timestamps{
					UpdatedAt: now.Add(-1 * time.Hour),
				},
			},
			LinkedIssues: []LinkedIssue{issue},
		},
		{
			Event: data.Event{
				Title: "Another PR for same issue",
				URL:   "https://github.com/org/repo/pull/2",
				Timestamps: data.Timestamps{
					UpdatedAt: now.Add(-2 * time.Hour),
				},
			},
			LinkedIssues: []LinkedIssue{issue},
		},
	}

	report := Aggregate(dateRange, prs, nil)

	assert.Empty(t, report.StandalonePRs)
	require.Len(t, report.IssueGroups, 1)
	assert.Equal(t, issue.Title, report.IssueGroups[0].Issue.Title)
	assert.Len(t, report.IssueGroups[0].PRs, 2)
}

func TestAggregate_MixedPRs(t *testing.T) {
	now := time.Now()
	dateRange := &DateRange{
		Start: now.Add(-24 * time.Hour),
		End:   now,
		Label: "2026-01-07",
	}

	issue := LinkedIssue{
		Number: 100,
		Title:  "Fix bug",
		URL:    "https://github.com/org/repo/issues/100",
	}

	prs := []PRWithIssues{
		{
			Event: data.Event{
				Title: "PR with issue",
				URL:   "https://github.com/org/repo/pull/1",
			},
			LinkedIssues: []LinkedIssue{issue},
		},
		{
			Event: data.Event{
				Title: "Standalone PR",
				URL:   "https://github.com/org/repo/pull/2",
			},
			LinkedIssues: nil,
		},
	}

	report := Aggregate(dateRange, prs, nil)

	assert.Len(t, report.IssueGroups, 1)
	assert.Len(t, report.StandalonePRs, 1)
}

func TestAggregate_Reviews(t *testing.T) {
	now := time.Now()
	dateRange := &DateRange{
		Start: now.Add(-24 * time.Hour),
		End:   now,
		Label: "2026-01-07",
	}

	reviews := []ReviewedPR{
		{
			Event: data.Event{
				Title:  "Someone's PR",
				URL:    "https://github.com/org/repo/pull/5",
				Repo:   "org/repo",
				Number: 5,
			},
			Reviews: []ReviewInfo{
				{State: "APPROVED", SubmittedAt: now.Add(-1 * time.Hour)},
				{State: "COMMENTED", SubmittedAt: now.Add(-2 * time.Hour)},
			},
		},
	}

	report := Aggregate(dateRange, nil, reviews)

	require.Len(t, report.ExtraReviews, 1)
	assert.Equal(t, "org", report.ExtraReviews[0].Owner)
	assert.Equal(t, "repo", report.ExtraReviews[0].Repo)
	assert.Equal(t, 5, report.ExtraReviews[0].PRNumber)
	assert.Len(t, report.ExtraReviews[0].Reviews, 2)
	assert.Equal(t, "COMMENTED", report.ExtraReviews[0].Reviews[0].State)
	assert.Equal(t, "APPROVED", report.ExtraReviews[0].Reviews[1].State)
}

func TestAggregate_SortsByActivityTime(t *testing.T) {
	now := time.Now()
	dateRange := &DateRange{
		Start: now.Add(-24 * time.Hour),
		End:   now,
		Label: "2026-01-07",
	}

	prs := []PRWithIssues{
		{
			Event: data.Event{
				Title: "Older PR",
				URL:   "https://github.com/org/repo/pull/1",
				Timestamps: data.Timestamps{
					UpdatedAt: now.Add(-5 * time.Hour),
				},
			},
		},
		{
			Event: data.Event{
				Title: "Newer PR",
				URL:   "https://github.com/org/repo/pull/2",
				Timestamps: data.Timestamps{
					UpdatedAt: now.Add(-1 * time.Hour),
				},
			},
		},
	}

	report := Aggregate(dateRange, prs, nil)

	require.Len(t, report.StandalonePRs, 2)
	assert.Equal(t, "Newer PR", report.StandalonePRs[0].Title)
	assert.Equal(t, "Older PR", report.StandalonePRs[1].Title)
}
