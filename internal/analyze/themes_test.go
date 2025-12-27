package analyze

import (
	"reflect"
	"testing"

	"github.com/jackchuka/gh-brag/internal/config"
	"github.com/jackchuka/gh-brag/internal/data"
)

func TestTheme(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Themes: []config.Theme{
			{Name: "Feature", Keywords: []string{"feat", "feature"}},
			{Name: "Bug Fix", Keywords: []string{"fix", "bug"}},
		},
	}
	analyzer, _ := New(cfg)

	tests := []struct {
		name     string
		events   []data.Event
		expected []Theme
	}{
		{
			name: "Match by label (highest priority)",
			events: []data.Event{
				{
					ID:     "1",
					Title:  "Some fix",
					Labels: []string{"feature"},
				},
			},
			expected: []Theme{
				{
					Name:  "Feature",
					Count: 1,
					Items: []data.Event{{ID: "1", Title: "Some fix", Labels: []string{"feature"}}},
				},
			},
		},
		{
			name: "Match by title (fallback)",
			events: []data.Event{
				{
					ID:    "2",
					Title: "Add new feat",
				},
			},
			expected: []Theme{
				{
					Name:  "Feature",
					Count: 1,
					Items: []data.Event{{ID: "2", Title: "Add new feat"}},
				},
			},
		},
		{
			name: "First appearance in title priority",
			events: []data.Event{
				{
					ID:    "3",
					Title: "fix: add feat",
				},
			},
			expected: []Theme{
				{
					Name:  "Bug Fix",
					Count: 1,
					Items: []data.Event{{ID: "3", Title: "fix: add feat"}},
				},
			},
		},
		{
			name: "Other category",
			events: []data.Event{
				{
					ID:    "4",
					Title: "random work",
				},
			},
			expected: []Theme{
				{
					Name:  "Other",
					Count: 1,
					Items: []data.Event{{ID: "4", Title: "random work"}},
				},
			},
		},
		{
			name: "Multiple events sorted by count",
			events: []data.Event{
				{ID: "1", Title: "fix 1"},
				{ID: "2", Title: "fix 2"},
				{ID: "3", Title: "feat 1"},
			},
			expected: []Theme{
				{
					Name:  "Bug Fix",
					Count: 2,
					Items: []data.Event{{ID: "1", Title: "fix 1"}, {ID: "2", Title: "fix 2"}},
				},
				{
					Name:  "Feature",
					Count: 1,
					Items: []data.Event{{ID: "3", Title: "feat 1"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := analyzer.theme(tt.events)

			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d themes, got %d", len(tt.expected), len(got))
			}

			for i := range got {
				if got[i].Name != tt.expected[i].Name {
					t.Errorf("expected theme %d name %s, got %s", i, tt.expected[i].Name, got[i].Name)
				}
				if got[i].Count != tt.expected[i].Count {
					t.Errorf("expected theme %d count %d, got %d", i, tt.expected[i].Count, got[i].Count)
				}
				if !reflect.DeepEqual(got[i].Items, tt.expected[i].Items) {
					t.Errorf("expected theme %d items %v, got %v", i, tt.expected[i].Items, got[i].Items)
				}
			}
		})
	}
}
