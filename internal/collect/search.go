package collect

import (
	"fmt"
	"time"

	"github.com/jackchuka/gh-brag/internal/data"
	"github.com/jackchuka/gh-brag/internal/github"
)

func RunSearch(kind string, action data.EventAction, query string) ([]data.Event, error) {
	nodes, err := github.RunSearch(query, github.QueryBasic)
	if err != nil {
		return nil, err
	}

	var events []data.Event
	fetchedAt := time.Now()

	// Map Kind to ID Prefix (prs -> pr, issues -> issue)
	var idPrefix string
	switch kind {
	case "prs":
		idPrefix = "pr"
	case "issues":
		idPrefix = "issue"
	default:
		idPrefix = kind
	}

	for _, n := range nodes {
		// Extract Labels
		labels := make([]string, 0, len(n.Labels.Nodes))
		for _, l := range n.Labels.Nodes {
			labels = append(labels, l.Name)
		}

		// Extract Reviewers (deduplicated)
		reviewersMap := make(map[string]bool)
		for _, r := range n.Reviews.Nodes {
			if r.Author.Login != "" {
				reviewersMap[r.Author.Login] = true
			}
		}
		var reviewers []string
		for r := range reviewersMap {
			reviewers = append(reviewers, r)
		}

		// Generate ID: prefix:url:action
		id := fmt.Sprintf("%s:%s:%s", idPrefix, n.URL, action)

		evt := data.Event{
			ID:        id,
			Action:    action,
			Kind:      kind,
			URL:       n.URL,
			Repo:      n.Repository.NameWithOwner,
			Number:    n.Number,
			Title:     n.Title,
			Body:      n.Body,
			Author:    n.Author.Login,
			Labels:    labels,
			Reviewers: reviewers,
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
		events = append(events, evt)
	}

	return events, nil
}
