package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackchuka/gh-brag/internal/analyze"
	"github.com/jackchuka/gh-brag/internal/config"
	"github.com/jackchuka/gh-brag/internal/data"
	"github.com/jackchuka/gh-brag/internal/visualize"
	"github.com/spf13/cobra"
)

var (
	visualizeIn string
)

var visualizeCmd = &cobra.Command{
	Use:   "visualize",
	Short: "Visualize your activity trends",
	Long:  `Displays a TUI dashboard showing your activity trends, impact, and collaboration.`,
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Open(visualizeIn)
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
		visualize.NewDashboard(metrics).Render()
	},
}

func init() {
	rootCmd.AddCommand(visualizeCmd)

	visualizeCmd.Flags().StringVar(&visualizeIn, "in", "gh-brag.events.jsonl", "Input JSONL file")
}
