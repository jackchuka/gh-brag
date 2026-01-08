package github

import (
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2"
)

// RunSearch executes a paginated GitHub GraphQL search and returns all matching nodes
func RunSearch(query string, queryType QueryType) ([]SearchNode, error) {
	var results []SearchNode
	cursor := ""
	graphqlQuery := GetQuery(queryType)

	for {
		args := []string{"api", "graphql", "-f", fmt.Sprintf("q=%s", query)}
		if cursor != "" {
			args = append(args, "-f", fmt.Sprintf("endCursor=%s", cursor))
		}
		args = append(args, "-f", fmt.Sprintf("query=%s", graphqlQuery))

		stdOut, _, err := gh.Exec(args...)
		if err != nil {
			return nil, fmt.Errorf("gh api failed: %w", err)
		}

		var resp searchResponse
		if err := json.Unmarshal(stdOut.Bytes(), &resp); err != nil {
			return nil, fmt.Errorf("failed to parse graphql response: %w", err)
		}

		results = append(results, resp.Data.Search.Nodes...)

		if !resp.Data.Search.PageInfo.HasNextPage {
			break
		}
		cursor = resp.Data.Search.PageInfo.EndCursor
	}

	return results, nil
}

// GetCurrentUser returns the authenticated GitHub username
func GetCurrentUser() (string, error) {
	stdOut, _, err := gh.Exec("api", "user", "-q", ".login")
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	// Trim newline
	login := stdOut.String()
	if len(login) > 0 && login[len(login)-1] == '\n' {
		login = login[:len(login)-1]
	}
	return login, nil
}
