package cmd

import (
	"fmt"
	"time"

	"github.com/jackchuka/gh-brag/internal/collect"
	"github.com/jackchuka/gh-brag/internal/data"
	"github.com/jackchuka/gh-brag/internal/spinner"
	"github.com/jackchuka/gh-brag/internal/store"
	"github.com/spf13/cobra"
)

var (
	collectFrom    string
	collectTo      string
	collectOut     string
	collectInclude string
	collectUser    string
	collectOwner   string
	collectRepo    string
)

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collects your activity from GitHub",
	Long:  `Searches GitHub for your PRs, Issues, and Reviews within a date range and saves them to a file.`,
	Run: func(cmd *cobra.Command, args []string) {
		s := spinner.NewSpinner(fmt.Sprintf(" Collecting data from %s to %s...", collectFrom, collectTo))
		s.Start()
		defer s.Stop()

		// Helper to print while spinner is active
		printInfo := func(msg string) {
			s.Stop()
			fmt.Println(msg)
			s.Start()
		}

		existingIDs, err := store.LoadExistingIDs(collectOut)
		if err != nil {
			printInfo(fmt.Sprintf("Error loading existing %s: %v", collectOut, err))
			return
		}

		var newEvents []data.Event

		// Helper to build query
		buildQuery := func(baseQuery string) string {
			q := baseQuery
			if collectOwner != "" {
				q += fmt.Sprintf(" user:%s", collectOwner)
			}
			if collectRepo != "" {
				q += fmt.Sprintf(" repo:%s", collectRepo)
			}
			return q
		}

		// 1. Authored PRs
		if collectInclude == "all" || collectInclude == "prs" {
			q := buildQuery(fmt.Sprintf("author:%s is:pr is:merged merged:%s..%s", collectUser, collectFrom, collectTo))
			s.Suffix = fmt.Sprintf(" Finding merged PRs (query: %s)...", q)

			res, err := collect.RunSearch("prs", data.EventActionMerged, q)
			if err != nil {
				printInfo(fmt.Sprintf("    Error: %v", err))
			} else {
				printInfo(fmt.Sprintf("    Found %d PRs", len(res)))
				for _, r := range res {
					if !existingIDs[r.ID] {
						newEvents = append(newEvents, r)
					}
				}
			}
		}

		// 2. Authored Issues
		if collectInclude == "all" || collectInclude == "issues" {
			q := buildQuery(fmt.Sprintf("author:%s is:issue created:%s..%s", collectUser, collectFrom, collectTo))
			s.Suffix = fmt.Sprintf(" Finding authored Issues (query: %s)...", q)

			res, err := collect.RunSearch("issues", data.EventActionAuthored, q)
			if err != nil {
				printInfo(fmt.Sprintf("    Error: %v", err))
			} else {
				printInfo(fmt.Sprintf("    Found %d Issues", len(res)))
				for _, r := range res {
					if !existingIDs[r.ID] {
						newEvents = append(newEvents, r)
					}
				}
			}
		}

		// 3. Reviewed PRs
		if collectInclude == "all" || collectInclude == "reviews" {
			q := buildQuery(fmt.Sprintf("is:pr reviewed-by:%s updated:%s..%s -author:%s", collectUser, collectFrom, collectTo, collectUser))
			s.Suffix = fmt.Sprintf(" Finding reviewed PRs (query: %s)...", q)

			res, err := collect.RunSearch("prs", data.EventActionReviewed, q)
			if err != nil {
				printInfo(fmt.Sprintf("    Error: %v", err))
			} else {
				printInfo(fmt.Sprintf("    Found %d Reviews", len(res)))
				for _, r := range res {
					if !existingIDs[r.ID] {
						newEvents = append(newEvents, r)
					}
				}
			}
		}

		s.Stop() // Stop spinner before final output

		if len(newEvents) > 0 {
			if err := store.AppendEvents(collectOut, newEvents); err != nil {
				fmt.Printf("Error saving events: %v\n", err)
			} else {
				fmt.Printf("Saved %d new events to %s\n", len(newEvents), collectOut)
			}
		} else {
			fmt.Println("No new events found.")
		}
	},
}

func init() {
	rootCmd.AddCommand(collectCmd)

	// Defaults
	now := time.Now()
	sixMonthsAgo := now.AddDate(0, -6, 0)
	defaultFrom := sixMonthsAgo.Format("2006-01-02")
	defaultTo := now.Format("2006-01-02")

	collectCmd.Flags().StringVar(&collectFrom, "from", defaultFrom, "Start date (YYYY-MM-DD)")
	collectCmd.Flags().StringVar(&collectTo, "to", defaultTo, "End date (YYYY-MM-DD)")
	collectCmd.Flags().StringVar(&collectOut, "out", "gh-brag.events.jsonl", "Output file path")
	collectCmd.Flags().StringVar(&collectInclude, "include", "all", "What to include: all, prs, issues, reviews")
	collectCmd.Flags().StringVar(&collectUser, "user", "@me", "GitHub username (optional, defaults to @me)")
	collectCmd.Flags().StringVar(&collectOwner, "owner", "", "Filter by owner (user or org)")
	collectCmd.Flags().StringVar(&collectRepo, "repo", "", "Filter by specific repository (e.g. owner/repo)")
}
