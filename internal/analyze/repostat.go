package analyze

import (
	"sort"

	"github.com/jackchuka/gh-brag/internal/data"
)

type RepoSummary struct {
	Name     string
	Merged   int
	Issues   int
	Reviewed int
}

type RepoStats struct {
	Summary []RepoSummary
}

func (a *Analyzer) repoStats(events []data.Event) RepoStats {
	issues := make(map[string]int)
	merged := make(map[string]int)
	reviewed := make(map[string]int)
	summary := make(map[string]RepoSummary)

	for _, e := range events {
		if e.Action == data.EventActionMerged {
			merged[e.Repo]++
		}
		if e.Action == data.EventActionReviewed {
			reviewed[e.Repo]++
		}
		if e.Action == data.EventActionAuthored {
			issues[e.Repo]++
		}
		summary[e.Repo] = RepoSummary{
			Name:     e.Repo,
			Merged:   merged[e.Repo],
			Issues:   issues[e.Repo],
			Reviewed: reviewed[e.Repo],
		}
	}

	return RepoStats{
		Summary: sortSummary(summary),
	}
}

func sortSummary(m map[string]RepoSummary) []RepoSummary {
	// merged desc, issues desc, reviewed desc
	var s []RepoSummary
	for _, v := range m {
		s = append(s, v)
	}
	sort.Slice(s, func(i, j int) bool {
		if s[i].Merged != s[j].Merged {
			return s[i].Merged > s[j].Merged
		}
		if s[i].Issues != s[j].Issues {
			return s[i].Issues > s[j].Issues
		}
		return s[i].Reviewed > s[j].Reviewed
	})
	return s
}
