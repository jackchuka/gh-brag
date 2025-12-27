package analyze

import (
	"reflect"
	"testing"

	"github.com/jackchuka/gh-brag/internal/config"
	"github.com/jackchuka/gh-brag/internal/data"
)

func TestCollaboration(t *testing.T) {
	t.Parallel()

	analyzer, _ := New(&config.Config{})

	tests := []struct {
		name     string
		events   []data.Event
		expected Collaboration
	}{
		{
			name: "Track reviewers and reviewees",
			events: []data.Event{
				{
					Action: data.EventActionReviewed,
					Author: "user-being-reviewed", // I reviewed this person
				},
				{
					Action:    data.EventActionMerged,
					Author:    "me",
					Reviewers: []string{"reviewer1", "reviewer2"},
				},
				{
					Action:    data.EventActionMerged,
					Author:    "me",
					Reviewers: []string{"reviewer1"},
				},
			},
			expected: Collaboration{
				Reviewers: []UserStat{
					{Login: "reviewer1", Count: 2},
					{Login: "reviewer2", Count: 1},
				},
				Reviewees: []UserStat{
					{Login: "user-being-reviewed", Count: 1},
				},
			},
		},
		{
			name: "Exclusion of self-reviews",
			events: []data.Event{
				{
					Action:    data.EventActionMerged,
					Author:    "me",
					Reviewers: []string{"me", "other"},
				},
			},
			expected: Collaboration{
				Reviewers: []UserStat{
					{Login: "other", Count: 1},
				},
				Reviewees: []UserStat{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := analyzer.collaboration(tt.events)

			if !reflect.DeepEqual(got.Reviewers, tt.expected.Reviewers) {
				t.Errorf("expected reviewers %v, got %v", tt.expected.Reviewers, got.Reviewers)
			}
			if !reflect.DeepEqual(got.Reviewees, tt.expected.Reviewees) {
				t.Errorf("expected reviewees %v, got %v", tt.expected.Reviewees, got.Reviewees)
			}
		})
	}
}
