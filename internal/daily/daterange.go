package daily

import (
	"fmt"
	"time"
)

// DateRange represents a time range for the daily report
type DateRange struct {
	Start time.Time
	End   time.Time
	Label string // YYYY-MM-DD or YYYY-MM-DD..YYYY-MM-DD
}

// ComputeRange calculates the date range based on the provided parameters.
// If date is provided, it returns that single day's range.
// If from/to are provided, it returns that range.
// If nothing is provided, it returns yesterday's range.
func ComputeRange(date, from, to, tzName string) (*DateRange, error) {
	loc, err := loadLocation(tzName)
	if err != nil {
		return nil, err
	}

	now := time.Now().In(loc)

	// Single date mode
	if date != "" {
		return computeSingleDate(date, loc)
	}

	// Range mode
	if from != "" && to != "" {
		return computeRangeFromTo(from, to, loc)
	}

	// Default: yesterday
	return computeYesterday(now, loc)
}

func loadLocation(tzName string) (*time.Location, error) {
	if tzName == "" {
		return time.Local, nil
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", tzName, err)
	}
	return loc, nil
}

func computeSingleDate(date string, loc *time.Location) (*DateRange, error) {
	d, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: %w", date, err)
	}

	start := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 1)

	return &DateRange{
		Start: start.UTC(),
		End:   end.UTC(),
		Label: date,
	}, nil
}

func computeRangeFromTo(from, to string, loc *time.Location) (*DateRange, error) {
	fromDate, err := time.ParseInLocation("2006-01-02", from, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid --from date %q: %w", from, err)
	}

	toDate, err := time.ParseInLocation("2006-01-02", to, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid --to date %q: %w", to, err)
	}

	if fromDate.After(toDate) {
		return nil, fmt.Errorf("--from date %q is after --to date %q", from, to)
	}

	start := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, loc)
	end := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)

	label := from
	if from != to {
		label = fmt.Sprintf("%s..%s", from, to)
	}

	return &DateRange{
		Start: start.UTC(),
		End:   end.UTC(),
		Label: label,
	}, nil
}

func computeYesterday(now time.Time, loc *time.Location) (*DateRange, error) {
	yesterday := now.AddDate(0, 0, -1)
	start := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 1)

	return &DateRange{
		Start: start.UTC(),
		End:   end.UTC(),
		Label: yesterday.Format("2006-01-02"),
	}, nil
}

// FormatStartForGitHub returns just the start date for open-ended queries (>= start)
func (r *DateRange) FormatStartForGitHub() string {
	return ">=" + r.Start.Format("2006-01-02T15:04:05Z")
}
