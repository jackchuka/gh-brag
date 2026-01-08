package daily

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeRange_SingleDate(t *testing.T) {
	tests := []struct {
		name        string
		date        string
		tz          string
		wantLabel   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid date",
			date:      "2026-01-07",
			wantLabel: "2026-01-07",
		},
		{
			name:      "valid date with timezone",
			date:      "2026-01-07",
			tz:        "America/New_York",
			wantLabel: "2026-01-07",
		},
		{
			name:        "invalid date format",
			date:        "01-07-2026",
			wantErr:     true,
			errContains: "invalid date",
		},
		{
			name:        "invalid timezone",
			date:        "2026-01-07",
			tz:          "Invalid/Zone",
			wantErr:     true,
			errContains: "invalid timezone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeRange(tt.date, "", "", tt.tz)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantLabel, result.Label)
			assert.True(t, result.End.After(result.Start))
			assert.Equal(t, 24*time.Hour, result.End.Sub(result.Start))
		})
	}
}

func TestComputeRange_FromTo(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		to          string
		wantLabel   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "same day",
			from:      "2026-01-07",
			to:        "2026-01-07",
			wantLabel: "2026-01-07",
		},
		{
			name:      "multi-day range",
			from:      "2026-01-01",
			to:        "2026-01-07",
			wantLabel: "2026-01-01..2026-01-07",
		},
		{
			name:        "from after to",
			from:        "2026-01-10",
			to:          "2026-01-05",
			wantErr:     true,
			errContains: "is after",
		},
		{
			name:        "invalid from date",
			from:        "invalid",
			to:          "2026-01-07",
			wantErr:     true,
			errContains: "invalid --from date",
		},
		{
			name:        "invalid to date",
			from:        "2026-01-01",
			to:          "invalid",
			wantErr:     true,
			errContains: "invalid --to date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeRange("", tt.from, tt.to, "")

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantLabel, result.Label)
			assert.True(t, result.End.After(result.Start))
		})
	}
}

func TestComputeRange_Default(t *testing.T) {
	result, err := ComputeRange("", "", "", "")

	require.NoError(t, err)
	assert.NotEmpty(t, result.Label)
	assert.True(t, result.End.After(result.Start))
	assert.Equal(t, 24*time.Hour, result.End.Sub(result.Start))
}

func TestDateRange_FormatStartForGitHub(t *testing.T) {
	loc, _ := time.LoadLocation("UTC")
	start := time.Date(2026, 1, 7, 0, 0, 0, 0, loc)

	dr := &DateRange{
		Start: start,
		End:   start.AddDate(0, 0, 1),
		Label: "2026-01-07",
	}

	result := dr.FormatStartForGitHub()
	assert.Equal(t, ">=2026-01-07T00:00:00Z", result)
}
