package daily

import (
	"testing"
	"time"

	"github.com/jackchuka/gh-brag/internal/data"
	"github.com/stretchr/testify/assert"
)

func TestActivityTime(t *testing.T) {
	now := time.Now()
	created := now.Add(-3 * time.Hour)
	updated := now.Add(-1 * time.Hour)
	closed := now.Add(-30 * time.Minute)

	tests := []struct {
		name     string
		event    data.Event
		expected time.Time
	}{
		{
			name: "merged PR returns ClosedAt",
			event: data.Event{
				Action: data.EventActionMerged,
				Timestamps: data.Timestamps{
					CreatedAt: created,
					UpdatedAt: updated,
					ClosedAt:  closed,
				},
			},
			expected: closed,
		},
		{
			name: "authored PR returns UpdatedAt",
			event: data.Event{
				Action: data.EventActionAuthored,
				Timestamps: data.Timestamps{
					CreatedAt: created,
					UpdatedAt: updated,
				},
			},
			expected: updated,
		},
		{
			name: "reviewed PR returns UpdatedAt",
			event: data.Event{
				Action: data.EventActionReviewed,
				Timestamps: data.Timestamps{
					CreatedAt: created,
					UpdatedAt: updated,
				},
			},
			expected: updated,
		},
		{
			name: "no UpdatedAt returns CreatedAt",
			event: data.Event{
				Action: data.EventActionAuthored,
				Timestamps: data.Timestamps{
					CreatedAt: created,
				},
			},
			expected: created,
		},
		{
			name: "merged with zero ClosedAt returns UpdatedAt",
			event: data.Event{
				Action: data.EventActionMerged,
				Timestamps: data.Timestamps{
					CreatedAt: created,
					UpdatedAt: updated,
				},
			},
			expected: updated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ActivityTime(tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}
