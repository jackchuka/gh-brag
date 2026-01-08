package llm

import (
	"fmt"
	"strings"
)

const systemPromptTemplate = `You are an AI assistant that summarizes a developer's daily GitHub activity.

Your task is to generate a concise summary of the developer's work based on their pull requests, issues, and code reviews.

The summary should be in %s and follow these guidelines:

Content:
- Use bullet points
- Be concise - one line per item
- Group by related activities into themes
- Read the titles and bodies for context
- Use simple language suitable for a general audience
- State the action taken in a sentence
- Each theme gets a top-level bullet with sub-bullets for details

Formatting:
- Add a blank line before and after code blocks
- Use backticks for code references, file names, and technical terms
- Ensure proper indentation for nested bullets (2 spaces)
- No trailing whitespace

DO NOT:
- Group by PRs or Reviews
- Write paragraphs
- Make up information not in the data
- Include URLs unless necessary
`

const maxBodyLength = 500

// langNames maps language codes to full names for better LLM understanding
var langNames = map[string]string{
	"en": "English",
	"ja": "Japanese",
}

// getLangName returns the full language name for a code
func getLangName(code string) string {
	if name, ok := langNames[code]; ok {
		return name
	}
	return code // fallback to code if not found
}

// truncateBody limits body text to avoid excessive prompt size
func truncateBody(body string) string {
	body = strings.TrimSpace(body)
	if len(body) > maxBodyLength {
		return body[:maxBodyLength] + "..."
	}
	return body
}

// buildMessages constructs the chat messages for the LLM
func buildMessages(cfg Config, input SummaryInput) []message {
	// Build structured user content
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("DATE: %s\n\n", input.DateLabel))

	// Issue groups (PRs organized by linked issues)
	if len(input.IssueGroups) > 0 {
		sb.WriteString("ISSUES WORKED ON:\n")
		for _, ig := range input.IssueGroups {
			sb.WriteString(fmt.Sprintf("- Issue: %s\n", ig.Title))
			if ig.Body != "" {
				sb.WriteString(fmt.Sprintf("  Description: %s\n", truncateBody(ig.Body)))
			}
			for _, pr := range ig.PRs {
				sb.WriteString(fmt.Sprintf("  - PR: %s\n", pr.Title))
				if pr.Body != "" {
					sb.WriteString(fmt.Sprintf("    Description: %s\n", truncateBody(pr.Body)))
				}
			}
		}
		sb.WriteString("\n")
	}

	// Standalone PRs (PRs without linked issues)
	if len(input.StandalonePRs) > 0 {
		sb.WriteString("AUTHORED PRs:\n")
		for _, pr := range input.StandalonePRs {
			sb.WriteString(fmt.Sprintf("- %s\n", pr.Title))
			if pr.Body != "" {
				sb.WriteString(fmt.Sprintf("  Description: %s\n", truncateBody(pr.Body)))
			}
		}
		sb.WriteString("\n")
	}

	// Reviews submitted
	if len(input.Reviews) > 0 {
		sb.WriteString("REVIEWS SUBMITTED:\n")
		for _, r := range input.Reviews {
			states := strings.Join(r.States, ", ")
			sb.WriteString(fmt.Sprintf("- %s on: %s\n", states, r.PRTitle))
		}
		sb.WriteString("\n")
	}

	// Add custom prompt if provided
	if cfg.Prompt != "" {
		sb.WriteString(fmt.Sprintf("\nAdditional instructions: %s\n", cfg.Prompt))
	}

	// Build system prompt with language
	langName := getLangName(cfg.Lang)
	systemPrompt := fmt.Sprintf(systemPromptTemplate, langName)

	return []message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: sb.String()},
	}
}
