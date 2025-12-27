package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootConfig string

var rootCmd = &cobra.Command{
	Use:   "gh-brag",
	Short: "A tool to aggregate GitHub receipts for your brag doc",
	Long: `gh-brag helps you collect your GitHub activity (PRs, Issues, Reviews)
and analyze it for high-value insights (ownership, collaboration)
via a TUI dashboard.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rootConfig, "config", "", "Path to configuration file (e.g., gh-brag-config.yaml)")
}
