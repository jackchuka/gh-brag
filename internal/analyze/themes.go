package analyze

import (
	"sort"
	"strings"

	"github.com/jackchuka/gh-brag/internal/data"
)

type Theme struct {
	Name  string
	Count int
	Items []data.Event
}

// theme groups events by theme based on keywords in their title.
func (a *Analyzer) theme(events []data.Event) []Theme {
	clusters := make(map[string][]data.Event)

	for _, e := range events {
		title := strings.ToLower(e.Title)
		matched := false
		// 1. Label Search (Highest Priority)
	LabelLoop:
		for _, label := range e.Labels {
			labelLower := strings.ToLower(label)
			for _, tm := range a.config.Themes {
				for _, k := range tm.Keywords {
					if strings.Contains(labelLower, strings.ToLower(k)) {
						clusters[tm.Name] = append(clusters[tm.Name], e)
						matched = true
						break LabelLoop
					}
				}
			}
		}

		// 2. Title Search (Secondary Fallback - First Appearance)
		if !matched {
			bestIndex := -1
			var bestTheme string

			for _, tm := range a.config.Themes {
				for _, k := range tm.Keywords {
					idx := strings.Index(title, strings.ToLower(k))
					if idx != -1 {
						if bestIndex == -1 || idx < bestIndex {
							bestIndex = idx
							bestTheme = tm.Name
						}
						if idx == 0 {
							break // Can't get better than the start
						}
					}
				}
			}

			if bestTheme != "" {
				clusters[bestTheme] = append(clusters[bestTheme], e)
				matched = true
			}
		}

		if !matched {
			clusters["Other"] = append(clusters["Other"], e)
		}
	}

	var themes []Theme
	for k, v := range clusters {
		themes = append(themes, Theme{Name: k, Count: len(v), Items: v})
	}

	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Count > themes[j].Count
	})

	return themes
}
