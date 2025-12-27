package collect

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cli/go-gh/v2"
	"github.com/jackchuka/gh-brag/internal/data"
)

type graphQLResponse struct {
	Data struct {
		Search struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []node `json:"nodes"`
		} `json:"search"`
	} `json:"data"`
}

type node struct {
	Typename   string `json:"__typename"`
	URL        string `json:"url"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ClosedAt  time.Time `json:"closedAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Labels struct {
		Nodes []struct {
			Name string `json:"name"`
		} `json:"nodes"`
	} `json:"labels"`
	Reviews struct {
		Nodes []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
		} `json:"nodes"`
	} `json:"reviews"`
}

const searchQuery = `
query($q: String!, $endCursor: String) {
	search(query: $q, type: ISSUE, first: 100, after: $endCursor) {
		pageInfo {
			hasNextPage
			endCursor
		}
		nodes {
			__typename
			... on PullRequest {
				url
				repository { nameWithOwner }
				number
				title
				body
				state
				createdAt
				updatedAt
				closedAt
				author { login }
				labels(first: 10) { nodes { name } }
				reviews(first: 10) { nodes { author { login } } }
			}
			... on Issue {
				url
				repository { nameWithOwner }
				number
				title
				body
				state
				createdAt
				updatedAt
				closedAt
				author { login }
				labels(first: 10) { nodes { name } }
			}
		}
	}
}`

func RunSearch(kind string, action data.EventAction, query string) ([]data.Event, error) {
	var events []data.Event
	fetchedAt := time.Now()
	cursor := ""

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

	// Loop for pagination
	for {
		args := []string{"api", "graphql", "-f", fmt.Sprintf("q=%s", query)}
		if cursor != "" {
			args = append(args, "-f", fmt.Sprintf("endCursor=%s", cursor))
		}
		args = append(args, "-f", fmt.Sprintf("query=%s", searchQuery))

		stdOut, _, err := gh.Exec(args...)
		if err != nil {
			return nil, fmt.Errorf("gh api failed: %w", err)
		}

		var resp graphQLResponse
		if err := json.Unmarshal(stdOut.Bytes(), &resp); err != nil {
			return nil, fmt.Errorf("failed to parse graphql response: %w", err)
		}

		for _, n := range resp.Data.Search.Nodes {
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

		if !resp.Data.Search.PageInfo.HasNextPage {
			break
		}
		cursor = resp.Data.Search.PageInfo.EndCursor
	}

	return events, nil
}
