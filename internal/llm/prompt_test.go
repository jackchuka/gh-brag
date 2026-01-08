package llm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLangName(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"en", "English"},
		{"ja", "Japanese"},
		{"fr", "fr"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := getLangName(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short text unchanged",
			input:    "short text",
			expected: "short text",
		},
		{
			name:     "whitespace trimmed",
			input:    "  text with spaces  ",
			expected: "text with spaces",
		},
		{
			name:     "long text truncated",
			input:    strings.Repeat("a", 600),
			expected: strings.Repeat("a", 500) + "...",
		},
		{
			name:     "exactly max length unchanged",
			input:    strings.Repeat("a", 500),
			expected: strings.Repeat("a", 500),
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateBody(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildMessages(t *testing.T) {
	tests := []struct {
		name            string
		cfg             Config
		input           SummaryInput
		wantSystemLang  string
		wantUserContent []string
	}{
		{
			name: "english with standalone PRs",
			cfg:  Config{Lang: "en"},
			input: SummaryInput{
				DateLabel: "2026-01-07",
				StandalonePRs: []PREntry{
					{Title: "Add feature", URL: "https://github.com/org/repo/pull/1"},
				},
			},
			wantSystemLang:  "English",
			wantUserContent: []string{"DATE: 2026-01-07", "AUTHORED PRs:", "Add feature"},
		},
		{
			name: "japanese language",
			cfg:  Config{Lang: "ja"},
			input: SummaryInput{
				DateLabel: "2026-01-07",
			},
			wantSystemLang:  "Japanese",
			wantUserContent: []string{"DATE: 2026-01-07"},
		},
		{
			name: "with issue groups",
			cfg:  Config{Lang: "en"},
			input: SummaryInput{
				DateLabel: "2026-01-07",
				IssueGroups: []IssueEntry{
					{
						Title: "Fix bug",
						URL:   "https://github.com/org/repo/issues/100",
						PRs: []PREntry{
							{Title: "Bug fix PR", URL: "https://github.com/org/repo/pull/1"},
						},
					},
				},
			},
			wantSystemLang:  "English",
			wantUserContent: []string{"ISSUES WORKED ON:", "Fix bug", "Bug fix PR"},
		},
		{
			name: "with reviews",
			cfg:  Config{Lang: "en"},
			input: SummaryInput{
				DateLabel: "2026-01-07",
				Reviews: []ReviewEntry{
					{PRTitle: "Someone's PR", PRURL: "https://github.com/org/repo/pull/5", States: []string{"APPROVED"}},
				},
			},
			wantSystemLang:  "English",
			wantUserContent: []string{"REVIEWS SUBMITTED:", "APPROVED on: Someone's PR"},
		},
		{
			name: "with custom prompt",
			cfg:  Config{Lang: "en", Prompt: "be formal"},
			input: SummaryInput{
				DateLabel: "2026-01-07",
			},
			wantSystemLang:  "English",
			wantUserContent: []string{"Additional instructions: be formal"},
		},
		{
			name: "with PR body",
			cfg:  Config{Lang: "en"},
			input: SummaryInput{
				DateLabel: "2026-01-07",
				StandalonePRs: []PREntry{
					{Title: "Add feature", URL: "https://github.com/org/repo/pull/1", Body: "This PR adds a new feature"},
				},
			},
			wantSystemLang:  "English",
			wantUserContent: []string{"Description: This PR adds a new feature"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := buildMessages(tt.cfg, tt.input)

			assert.Len(t, messages, 2)
			assert.Equal(t, "system", messages[0].Role)
			assert.Equal(t, "user", messages[1].Role)

			assert.Contains(t, messages[0].Content, tt.wantSystemLang)

			for _, want := range tt.wantUserContent {
				assert.Contains(t, messages[1].Content, want)
			}
		})
	}
}
