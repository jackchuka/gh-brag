package github

// QueryType indicates which GraphQL query variant to use
type QueryType int

const (
	// QueryBasic fetches PR/Issue with labels and reviewer logins (for collect command)
	QueryBasic QueryType = iota
	// QueryWithLinkedIssues fetches PR with closingIssuesReferences and mergedAt (for daily authored PRs)
	QueryWithLinkedIssues
	// QueryWithReviews fetches PR with review details including state and submittedAt (for daily reviews)
	QueryWithReviews
)

// queryBasic is for the collect command - includes labels and reviewer logins
const queryBasic = `
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

// queryWithLinkedIssues is for daily authored PRs - includes closingIssuesReferences and mergedAt
const queryWithLinkedIssues = `
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
				mergedAt
				author { login }
				labels(first: 10) { nodes { name } }
				closingIssuesReferences(first: 10) {
					nodes {
						number
						title
						url
					}
				}
			}
		}
	}
}`

// queryWithReviews is for daily reviewed PRs - includes review details
const queryWithReviews = `
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
				createdAt
				updatedAt
				closedAt
				author { login }
				reviews(first: 100) {
					nodes {
						state
						submittedAt
						url
						author { login }
					}
				}
			}
		}
	}
}`

// GetQuery returns the GraphQL query string for the given query type
func GetQuery(qt QueryType) string {
	switch qt {
	case QueryWithLinkedIssues:
		return queryWithLinkedIssues
	case QueryWithReviews:
		return queryWithReviews
	default:
		return queryBasic
	}
}
