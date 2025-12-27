package analyze

import (
	"testing"
	"time"

	"github.com/jackchuka/gh-brag/internal/config"
	"github.com/jackchuka/gh-brag/internal/data"
)

func TestAnalyze(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Themes: []config.Theme{
			{Name: "Feature", Keywords: []string{"feat"}},
		},
		Metrics: config.Metrics{
			OwnershipThreshold: 2,
			ActionWeights: map[data.EventAction]float64{
				data.EventActionMerged:   10.0,
				data.EventActionAuthored: 5.0,
			},
			ThemeWeights: map[string]float64{
				"Feature": 2.0,
			},
		},
	}
	analyzer, _ := New(cfg)

	now := time.Now().UTC()
	isoYear, isoWeek := now.ISOWeek()
	startOfWeek := getStartOfWeek(isoYear, isoWeek)

	tests := []struct {
		name     string
		events   []data.Event
		validate func(t *testing.T, m Metrics)
	}{
		{
			name:   "Empty events",
			events: []data.Event{},
			validate: func(t *testing.T, m Metrics) {
				if m.Version != "" {
					t.Errorf("expected empty metrics, got version %s", m.Version)
				}
			},
		},
		{
			name: "Impact score calculation",
			events: []data.Event{
				{
					ID:     "1",
					Action: data.EventActionMerged, // weight 10
					Title:  "feat: something",      // theme Feature -> weight 2
					Timestamps: data.Timestamps{
						UpdatedAt: now,
					},
				},
				{
					ID:     "2",
					Action: data.EventActionAuthored, // weight 5
					Title:  "random",                 // theme Other -> weight 1
					Timestamps: data.Timestamps{
						UpdatedAt: now,
					},
				},
			},
			validate: func(t *testing.T, m Metrics) {
				// (10 * 2) + (5 * 1) = 25
				if m.ImpactScore != 25.0 {
					t.Errorf("expected impact score 25.0, got %f", m.ImpactScore)
				}
			},
		},
		{
			name: "Ownership count threshold",
			events: []data.Event{
				{ID: "1", Action: data.EventActionMerged, Repo: "org/repo1", Title: "fix", Timestamps: data.Timestamps{UpdatedAt: now}},
				{ID: "2", Action: data.EventActionMerged, Repo: "org/repo1", Title: "fix", Timestamps: data.Timestamps{UpdatedAt: now}},
				{ID: "3", Action: data.EventActionMerged, Repo: "org/repo2", Title: "fix", Timestamps: data.Timestamps{UpdatedAt: now}},
			},
			validate: func(t *testing.T, m Metrics) {
				// repo1 has 2 merged, threshold is 2 -> count 1
				if m.OwnershipCount != 1 {
					t.Errorf("expected ownership count 1, got %d", m.OwnershipCount)
				}
			},
		},
		{
			name: "Weekly trend and velocity",
			events: []data.Event{
				{ID: "1", Title: "fix", Timestamps: data.Timestamps{UpdatedAt: startOfWeek}},
				{ID: "2", Title: "fix", Timestamps: data.Timestamps{UpdatedAt: startOfWeek}},
				{ID: "3", Title: "fix", Timestamps: data.Timestamps{UpdatedAt: startOfWeek.AddDate(0, 0, 7)}}, // Next week
			},
			validate: func(t *testing.T, m Metrics) {
				if len(m.WeeklyTrend) != 2 {
					t.Fatalf("expected 2 trend points, got %d", len(m.WeeklyTrend))
				}
				if m.WeeklyTrend[0].Count != 2 {
					t.Errorf("expected week 1 count 2, got %d", m.WeeklyTrend[0].Count)
				}
				if m.WeeklyTrend[1].Count != 1 {
					t.Errorf("expected week 2 count 1, got %d", m.WeeklyTrend[1].Count)
				}
				// Velocity: 3 events over ~1 week = 3 (actually min 1 week)
				if m.Velocity != 3.0 {
					t.Errorf("expected velocity 3.0, got %f", m.Velocity)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := analyzer.Analyze(tt.events)
			tt.validate(t, got)
		})
	}
}
