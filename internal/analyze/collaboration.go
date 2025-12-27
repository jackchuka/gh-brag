package analyze

import (
	"sort"

	"github.com/jackchuka/gh-brag/internal/data"
)

type UserStat struct {
	Login string
	Count int
}

type Collaboration struct {
	Reviewers []UserStat // People who reviewed me
	Reviewees []UserStat // People I reviewed
}

func (a *Analyzer) collaboration(events []data.Event) Collaboration {
	reviewees := make(map[string]int)
	reviewersMap := make(map[string]int)

	for _, e := range events {
		if e.Action == data.EventActionReviewed {
			reviewees[e.Author]++
		}

		if e.Action == data.EventActionMerged {
			for _, reviewer := range e.Reviewers {
				if reviewer != e.Author { // Skip self-reviews
					reviewersMap[reviewer]++
				}
			}
		}
	}

	return Collaboration{
		Reviewers: sortStats(reviewersMap),
		Reviewees: sortStats(reviewees),
	}
}

func sortStats(m map[string]int) []UserStat {
	var s []UserStat
	for k, v := range m {
		s = append(s, UserStat{Login: k, Count: v})
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i].Count > s[j].Count // Descending
	})
	return s
}
