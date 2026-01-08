package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackchuka/gh-brag/internal/daily"
	"github.com/jackchuka/gh-brag/internal/github"
	"github.com/jackchuka/gh-brag/internal/llm"
	"github.com/jackchuka/gh-brag/internal/spinner"
	"github.com/spf13/cobra"
)

var (
	dailyDate                string
	dailyFrom                string
	dailyTo                  string
	dailyTz                  string
	dailyFormat              string
	dailyIncludeLinkedIssues bool
	dailyIncludeReviews      bool
	dailyOrgs                []string

	// Summarization flags
	dailySummarize        bool
	dailySummarizeLang    string
	dailySummarizeModel   string
	dailySummarizePrompt  string
	dailySummarizeTimeout time.Duration
)

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Generate a daily activity report",
	Long:  `Fetches your GitHub PR activity and generates a report for a specific day or date range.`,
	RunE:  runDaily,
}

func init() {
	rootCmd.AddCommand(dailyCmd)

	dailyCmd.Flags().StringVar(&dailyDate, "date", "", "Report date (YYYY-MM-DD)")
	dailyCmd.Flags().StringVar(&dailyFrom, "from", "", "Range start (YYYY-MM-DD)")
	dailyCmd.Flags().StringVar(&dailyTo, "to", "", "Range end (YYYY-MM-DD)")
	dailyCmd.Flags().StringVar(&dailyTz, "tz", "", "Timezone (IANA name, e.g., America/New_York)")
	dailyCmd.Flags().StringVar(&dailyFormat, "format", "plain", "Output format: json, yaml, plain")
	dailyCmd.Flags().BoolVar(&dailyIncludeLinkedIssues, "include-linked-issues", true, "Include linked issues")
	dailyCmd.Flags().BoolVar(&dailyIncludeReviews, "include-reviews", true, "Include submitted reviews")
	dailyCmd.Flags().StringSliceVar(&dailyOrgs, "org", nil, "Filter by organization(s), repeatable")

	// Summarization flags
	dailyCmd.Flags().BoolVar(&dailySummarize, "summarize", false, "Generate LLM summary using GitHub Models")
	dailyCmd.Flags().StringVar(&dailySummarizeLang, "summarize-lang", "en", "Output language (en, ja, etc.)")
	dailyCmd.Flags().StringVar(&dailySummarizeModel, "summarize-model", "openai/gpt-4o", "Model name")
	dailyCmd.Flags().StringVar(&dailySummarizePrompt, "summarize-prompt", "", "Additional prompt instructions")
	dailyCmd.Flags().DurationVar(&dailySummarizeTimeout, "summarize-timeout", 30*time.Second, "Request timeout")
}

func runDaily(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateDailyFlags(); err != nil {
		return err
	}

	// Validate format
	if dailyFormat != "json" && dailyFormat != "yaml" && dailyFormat != "plain" {
		return fmt.Errorf("invalid format %q: must be json, yaml, or plain", dailyFormat)
	}

	// Compute date range
	dateRange, err := daily.ComputeRange(dailyDate, dailyFrom, dailyTo, dailyTz)
	if err != nil {
		return fmt.Errorf("failed to compute date range: %w", err)
	}

	// Start spinner
	s := spinner.NewSpinner(fmt.Sprintf(" Fetching activity for %s...", dateRange.Label))
	s.Start()

	// Get current user for filtering reviews
	currentUser, err := github.GetCurrentUser()
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// Fetch authored PRs
	s.Suffix = " Fetching authored PRs..."
	prs, err := daily.FetchAuthoredPRs(dateRange, dailyIncludeLinkedIssues, dailyOrgs)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to fetch PRs: %w", err)
	}

	// Fetch reviewed PRs
	var reviews []daily.ReviewedPR
	if dailyIncludeReviews {
		s.Suffix = " Fetching reviews..."
		reviews, err = daily.FetchReviewedPRs(dateRange, currentUser, dailyOrgs)
		if err != nil {
			s.Stop()
			return fmt.Errorf("failed to fetch reviews: %w", err)
		}
	}

	// Aggregate results
	report := daily.Aggregate(dateRange, prs, reviews)

	// Generate summary if requested
	if dailySummarize {
		s.Suffix = " Generating summary..."
		summary, err := summarizeReport(report)
		if err != nil {
			s.Stop()
			fmt.Fprintf(os.Stderr, "Warning: summarization failed: %v\n", err)
		} else {
			report.Summary = summary
		}
	}

	s.Stop()

	// Render output
	output, err := daily.Render(report, dailyFormat)
	if err != nil {
		return fmt.Errorf("failed to render report: %w", err)
	}

	fmt.Println(output)
	return nil
}

func validateDailyFlags() error {
	// Check for mutually exclusive flags
	if dailyDate != "" && (dailyFrom != "" || dailyTo != "") {
		return errors.New("cannot use --date with --from/--to")
	}

	// Check that --from and --to are used together
	if (dailyFrom != "" && dailyTo == "") || (dailyFrom == "" && dailyTo != "") {
		return errors.New("--from and --to must be used together")
	}

	return nil
}

// summarizeReport generates an LLM summary of the daily report
func summarizeReport(report *daily.DailyReport) (string, error) {
	cfg := llm.Config{
		Model:   dailySummarizeModel,
		Lang:    dailySummarizeLang,
		Prompt:  dailySummarizePrompt,
		Timeout: dailySummarizeTimeout,
	}

	input := llm.SummaryInput{
		DateLabel: report.DateLabel,
	}

	// Convert issue groups
	for _, ig := range report.IssueGroups {
		entry := llm.IssueEntry{
			Title: ig.Issue.Title,
			URL:   ig.Issue.URL,
		}
		for _, pr := range ig.PRs {
			entry.PRs = append(entry.PRs, llm.PREntry{
				Title: pr.Title,
				URL:   pr.URL,
				Body:  pr.Body,
			})
		}
		input.IssueGroups = append(input.IssueGroups, entry)
	}

	// Convert standalone PRs
	for _, pr := range report.StandalonePRs {
		input.StandalonePRs = append(input.StandalonePRs, llm.PREntry{
			Title: pr.Title,
			URL:   pr.URL,
			Body:  pr.Body,
		})
	}

	// Convert reviews
	for _, r := range report.ExtraReviews {
		states := make([]string, len(r.Reviews))
		for i, rev := range r.Reviews {
			states[i] = rev.State
		}
		input.Reviews = append(input.Reviews, llm.ReviewEntry{
			PRTitle: r.PRTitle,
			PRURL:   r.PRURL,
			States:  states,
		})
	}

	return llm.Summarize(context.Background(), cfg, input)
}
