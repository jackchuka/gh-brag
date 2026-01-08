package github

import "time"

// SearchNode represents a unified node from GitHub GraphQL search
// covering all fields needed by collect and daily commands
type SearchNode struct {
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
	MergedAt  time.Time `json:"mergedAt"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Labels struct {
		Nodes []LabelNode `json:"nodes"`
	} `json:"labels"`
	Reviews struct {
		Nodes []ReviewNode `json:"nodes"`
	} `json:"reviews"`
	ClosingIssuesReferences struct {
		Nodes []LinkedIssueNode `json:"nodes"`
	} `json:"closingIssuesReferences"`
}

// LabelNode represents a label on a PR/Issue
type LabelNode struct {
	Name string `json:"name"`
}

// ReviewNode represents a review on a PR
type ReviewNode struct {
	State       string    `json:"state"`
	SubmittedAt time.Time `json:"submittedAt"`
	URL         string    `json:"url"`
	Author      struct {
		Login string `json:"login"`
	} `json:"author"`
}

// LinkedIssueNode represents an issue linked via closingIssuesReferences
type LinkedIssueNode struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

// searchResponse is the internal response structure for GraphQL search
type searchResponse struct {
	Data struct {
		Search struct {
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			Nodes []SearchNode `json:"nodes"`
		} `json:"search"`
	} `json:"data"`
}
