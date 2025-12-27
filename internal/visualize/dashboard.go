package visualize

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jackchuka/gh-brag/internal/analyze"
)

type dashboard struct {
	metrics analyze.Metrics
}

func NewDashboard(metrics analyze.Metrics) *dashboard {
	return &dashboard{
		metrics: metrics,
	}
}

var (
	primaryColor = lipgloss.Color("#7D56F4") // Deep Purple
	successColor = lipgloss.Color("#00C094") // Emerald
	alertColor   = lipgloss.Color("#FF4672") // Hot Pink
	neutralColor = lipgloss.Color("#5A5A5A") // Slate Grey
)

func (d *dashboard) Render() {
	// 1. Hero Banner
	d.renderHero()

	// 2. Metrics Bar (KPIs)
	d.renderKPIs()

	// 3. Main Content Grid (Themes & Repos)
	left := d.renderThemeDist()
	right := d.renderRepoPulse()
	fmt.Println(lipgloss.JoinHorizontal(lipgloss.Top, left, right))

	// 4. Activity Pulse (Heatmap)
	d.renderHeatmap()

	// 5. Collaboration Network
	d.renderCollabNetwork()
}

func (d *dashboard) renderHero() {
	heroStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		MarginBottom(1).
		Width(80)

	title := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("🚀 GH-BRAG DASHBOARD")
	period := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("%s - %s",
		d.metrics.PeriodStart.Format("Jan 2006"),
		d.metrics.PeriodEnd.Format("Jan 2006")))

	summary := d.generateSummary()
	fmt.Println(heroStyle.Render(fmt.Sprintf("%s %s\n\n%s", title, period, summary)))
}

func (d *dashboard) generateSummary() string {
	if len(d.metrics.Theme) == 0 {
		return "You've been quiet this period. Start some work to see insights!"
	}

	topTheme := d.metrics.Theme[0].Name
	topRepo := ""
	if len(d.metrics.RepoStats.Summary) > 0 {
		topRepo = d.metrics.RepoStats.Summary[0].Name
	}

	impactLevel := "Active"
	if d.metrics.ImpactScore > 500 {
		impactLevel = "Powerhouse"
	} else if d.metrics.ImpactScore > 200 {
		impactLevel = "Driver"
	}

	return fmt.Sprintf("You've been a %s %s this period, with significant impact in %s.",
		lipgloss.NewStyle().Italic(true).Render(topTheme),
		lipgloss.NewStyle().Italic(true).Render(impactLevel),
		lipgloss.NewStyle().Italic(true).Render(topRepo))
}

func (d *dashboard) renderKPIs() {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(neutralColor).
		Padding(1, 2).
		MarginRight(2).
		Width(24).
		Align(lipgloss.Center)

	impact := cardStyle.Render(fmt.Sprintf("IMPACT\n%s\n%s",
		lipgloss.NewStyle().Bold(true).Foreground(alertColor).Render(fmt.Sprintf("%.1f", d.metrics.ImpactScore)),
		lipgloss.NewStyle().Faint(true).Render("⚡ High Power")))

	velocity := cardStyle.Render(fmt.Sprintf("VELOCITY\n%s\n%s",
		lipgloss.NewStyle().Bold(true).Foreground(successColor).Render(fmt.Sprintf("%.1f", d.metrics.Velocity)),
		lipgloss.NewStyle().Faint(true).Render("🔥 ev/wk")))

	ownership := cardStyle.Render(fmt.Sprintf("OWNERSHIP\n%s\n%s",
		lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(fmt.Sprintf("%d", d.metrics.OwnershipCount)),
		lipgloss.NewStyle().Faint(true).Render("🏆 Core Repos")))

	fmt.Println(lipgloss.JoinHorizontal(lipgloss.Top, impact, velocity, ownership))
	fmt.Println()
}

func (d *dashboard) renderThemeDist() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("THEME DISTRIBUTION")

	var content strings.Builder
	content.WriteString(title + "\n\n")

	sort.Slice(d.metrics.Theme, func(i, j int) bool { return d.metrics.Theme[i].Count > d.metrics.Theme[j].Count })

	maxCount := 0
	for _, t := range d.metrics.Theme {
		if t.Count > maxCount {
			maxCount = t.Count
		}
	}

	for _, t := range d.metrics.Theme {
		if t.Count == 0 {
			continue
		}
		barWidth := 15
		filled := (t.Count * barWidth) / maxCount
		bar := lipgloss.NewStyle().Foreground(successColor).Render(strings.Repeat("█", filled))
		empty := lipgloss.NewStyle().Foreground(neutralColor).Render(strings.Repeat("░", barWidth-filled))

		name := t.Name
		if len(name) > 12 {
			name = name[:9] + "..."
		}

		content.WriteString(fmt.Sprintf("%-12s %s%s %d\n", name, bar, empty, t.Count))
	}

	return lipgloss.NewStyle().Padding(1, 2).Width(42).Render(content.String())
}

func (d *dashboard) renderRepoPulse() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("TOP REPOSITORIES")
	var content strings.Builder
	content.WriteString(title + "\n\n")

	maxMerged := 0
	for _, r := range d.metrics.RepoStats.Summary {
		if r.Merged > maxMerged {
			maxMerged = r.Merged
		}
	}

	for i, r := range d.metrics.RepoStats.Summary {
		if i >= 10 {
			break
		}
		pulseWidth := 10
		filled := 0
		if maxMerged > 0 {
			filled = (r.Merged * pulseWidth) / maxMerged
		}
		pulse := lipgloss.NewStyle().Foreground(alertColor).Render(strings.Repeat("█", filled)) +
			lipgloss.NewStyle().Foreground(neutralColor).Render(strings.Repeat("▒", pulseWidth-filled))

		name := r.Name
		if len(name) > 18 {
			name = name[:15] + "..."
		}
		content.WriteString(fmt.Sprintf("%-18s %s %2d\n", name, pulse, r.Merged))
	}

	return lipgloss.NewStyle().Padding(1, 1).Width(40).Render(content.String())
}

func (d *dashboard) renderHeatmap() {
	title := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("ACTIVITY INTENSITY")
	intensity := fmt.Sprintf("%s %s %s",
		lipgloss.NewStyle().Foreground(alertColor).Render("█")+" High",
		lipgloss.NewStyle().Foreground(successColor).Render("▓")+" Medium",
		lipgloss.NewStyle().Foreground(primaryColor).Render("▒")+" Low",
	)

	fmt.Println(title, intensity)

	if len(d.metrics.WeeklyTrend) == 0 {
		fmt.Println("No activity data available.")
		return
	}

	const wrapAt = 13
	for i := 0; i < len(d.metrics.WeeklyTrend); i += wrapAt {
		endIdx := min(i+wrapAt, len(d.metrics.WeeklyTrend))
		rowWeeks := d.metrics.WeeklyTrend[i:endIdx]

		var header string
		var row string
		for _, w := range rowWeeks {
			t, _ := time.Parse("2006-01-02", w.Date)
			monthLabel := ""
			if t.Day() <= 7 {
				monthLabel = t.Format("Jan")
			}
			monthStr := fmt.Sprintf("%-5s", monthLabel)
			header += lipgloss.NewStyle().Faint(true).Render(monthStr)

			val := w.Count
			char := "░"
			color := neutralColor
			if val > 40 {
				char = "█"
				color = alertColor
			} else if val > 20 {
				char = "▓"
				color = successColor
			} else if val > 0 {
				char = "▒"
				color = primaryColor
			}
			row += fmt.Sprintf("[%s]  ", lipgloss.NewStyle().Foreground(color).Render(char))
		}
		fmt.Println(header)
		fmt.Println(row)
		fmt.Println()
	}
}

func (d *dashboard) renderCollabNetwork() {
	title := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("COLLABORATION NETWORK")
	fmt.Println(title)

	leftCol := lipgloss.NewStyle().Width(38).Render(d.renderUserList("Review Council", d.metrics.Collaboration.Reviewers))
	rightCol := lipgloss.NewStyle().Width(38).Render(d.renderUserList("Mentorship Impact", d.metrics.Collaboration.Reviewees))

	fmt.Println(lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol))
	fmt.Println()
}

func (d *dashboard) renderUserList(title string, users []analyze.UserStat) string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Underline(true).Render(title) + "\n")
	for i, u := range users {
		if i >= 5 {
			break
		}
		login := u.Login
		if len(login) > 15 {
			login = login[:12] + "..."
		}
		sb.WriteString(fmt.Sprintf(" 👤 %-15s [%2d]\n", login, u.Count))
	}
	return sb.String()
}
