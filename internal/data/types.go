package data

import "time"

type EventAction string

const (
	EventActionMerged   EventAction = "merged"
	EventActionReviewed EventAction = "reviewed"
	EventActionAuthored EventAction = "authored"
)

type Event struct {
	ID     string      `json:"id"`     // Unique ID: kind:url:action
	Action EventAction `json:"action"` // "merged", "reviewed", "authored"
	Kind   string      `json:"kind"`   // "pr", "issue"

	URL       string   `json:"url"`
	Repo      string   `json:"repo"`
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Author    string   `json:"author"`
	Labels    []string `json:"labels,omitempty"`
	Reviewers []string `json:"reviewers"` // List of reviewer logins

	Timestamps Timestamps `json:"timestamps"`
	Source     Source     `json:"source"`
}

type Timestamps struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ClosedAt  time.Time `json:"closedAt"`
}

type Source struct {
	Tool      string    `json:"tool"`
	Query     string    `json:"query"`
	FetchedAt time.Time `json:"fetchedAt"`
}
