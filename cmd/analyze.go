package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackchuka/gh-brag/internal/analyze"
	"github.com/jackchuka/gh-brag/internal/config"
	"github.com/jackchuka/gh-brag/internal/data"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	analyzeIn  string
	analyzeOut string
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze collected data for insights",
	Long:  `Generates a detailed YAML report containing metrics, theme clusters, and collaboration insights.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Analyzing data from %s...\n", analyzeIn)

		f, err := os.Open(analyzeIn)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer func() {
			_ = f.Close()
		}()

		var events []data.Event
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var e data.Event
			if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
				events = append(events, e)
			}
		}

		cfg, err := config.LoadConfig(rootConfig)
		if err != nil {
			fmt.Printf("Warning: error loading config: %v. Using defaults.\n", err)
			cfg = &config.Config{}
		}

		analyzer, err := analyze.New(cfg)
		if err != nil {
			fmt.Printf("Error creating analyzer: %v\n", err)
			return
		}
		metrics := analyzer.Analyze(events)

		// Marshal to YAML
		out, err := yaml.Marshal(metrics)
		if err != nil {
			fmt.Printf("Error marshaling metrics to YAML: %v\n", err)
			return
		}

		if err := os.WriteFile(analyzeOut, out, 0644); err != nil {
			fmt.Printf("Error writing report: %v\n", err)
			return
		}
		fmt.Printf("Analysis complete. Report written to %s\n", analyzeOut)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVar(&analyzeIn, "in", "gh-brag.events.jsonl", "Input JSONL file")
	analyzeCmd.Flags().StringVar(&analyzeOut, "out", "gh-brag.report.yaml", "Output report file")
}
