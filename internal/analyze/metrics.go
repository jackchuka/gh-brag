package analyze

import (
	"time"

	"github.com/jackchuka/gh-brag/internal/data"
)

// TrendPoint represents activity count for a specific date.
type TrendPoint struct {
	Date  string
	Count int
}

// Metrics are advanced metrics computed from events.
type Metrics struct {
	// Version
	Version string

	// Raw metrics
	RepoStats     RepoStats
	Theme         []Theme
	Collaboration Collaboration

	// Derived metrics
	PeriodStart     time.Time
	PeriodEnd       time.Time
	ImpactScore     float64            // Sum of action weights * theme weights
	Velocity        float64            // events per week
	OwnershipCount  int                // Number of repos with >= ownership threshold
	WeeklyTrend     []TrendPoint       // Number of events per week (ordered)
	ContributionMix map[string]float64 // Percentage of events per theme
}

// Analyze computes all advanced metrics and analysis from the given events.
func (a *Analyzer) Analyze(events []data.Event) Metrics {
	if len(events) == 0 {
		return Metrics{}
	}

	report := Metrics{
		Version:         "1.0.0",
		ContributionMix: make(map[string]float64),
	}

	// Aggregate Raw Metrics
	report.Theme = a.theme(events)
	report.RepoStats = a.repoStats(events)
	report.Collaboration = a.collaboration(events)

	// Theme Clusters (What you worked on)
	total := float64(len(events))
	themeMap := make(map[string]string)
	for _, t := range report.Theme {
		report.ContributionMix[t.Name] = (float64(t.Count) / total) * 100
		for _, e := range t.Items {
			themeMap[e.ID] = t.Name
		}
	}

	// Impact Score Calculation
	var totalImpact float64
	for _, e := range events {
		// Action Weight
		weight := a.config.Metrics.ActionWeights[e.Action]
		if weight == 0 {
			weight = 1.0 // Default if unknown
		}

		// Theme Weight
		theme := themeMap[e.ID]
		multiplier := a.config.Metrics.ThemeWeights[theme]
		if multiplier == 0 {
			multiplier = 1.0
		}
		totalImpact += weight * multiplier
	}
	report.ImpactScore = totalImpact

	// Velocity and Trend Calculation
	var minDate, maxDate time.Time
	trendAgg := make(map[string]int)
	for _, e := range events {
		t := e.Timestamps.UpdatedAt
		if minDate.IsZero() || t.Before(minDate) {
			minDate = t
		}
		if maxDate.IsZero() || t.After(maxDate) {
			maxDate = t
		}

		year, week := t.ISOWeek()
		key := getStartOfWeek(year, week).Format("2006-01-02")
		trendAgg[key]++
	}

	report.PeriodStart = minDate
	report.PeriodEnd = maxDate

	if !minDate.IsZero() {
		minYear, minWeek := minDate.ISOWeek()
		maxYear, maxWeek := maxDate.ISOWeek()
		current := getStartOfWeek(minYear, minWeek)
		end := getStartOfWeek(maxYear, maxWeek)

		for !current.After(end) {
			key := current.Format("2006-01-02")
			report.WeeklyTrend = append(report.WeeklyTrend, TrendPoint{
				Date:  key,
				Count: trendAgg[key],
			})
			current = current.AddDate(0, 0, 7)
		}
	}

	weeks := maxDate.Sub(minDate).Hours() / 24 / 7
	if weeks < 1 {
		weeks = 1
	}
	report.Velocity = float64(len(events)) / weeks

	// Ownership Index
	for _, c := range report.RepoStats.Summary {
		if c.Merged >= a.config.Metrics.OwnershipThreshold {
			report.OwnershipCount++
		}
	}

	// Collaboration Graph (Who you work with)
	report.Collaboration = a.collaboration(events)

	return report
}

// getStartOfWeek returns the Monday of the given ISO year and week.
func getStartOfWeek(year, week int) time.Time {
	// Jan 4th is always in ISO week 1
	t := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)
	// Roll back to Monday
	daysToRollBack := int(t.Weekday()) - 1
	if daysToRollBack < 0 {
		daysToRollBack = 6 // Sunday
	}
	t = t.AddDate(0, 0, -daysToRollBack)
	// Add (week-1) weeks
	return t.AddDate(0, 0, (week-1)*7)
}
