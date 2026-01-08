package daily

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed plain.tmpl
var plainTemplate string

// RenderPlain renders the report as plain text using the embedded template
func RenderPlain(report *DailyReport) (string, error) {
	tmpl, err := template.New("plain").Parse(plainTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, report); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderJSON renders the report as JSON
func RenderJSON(report *DailyReport) ([]byte, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

// RenderYAML renders the report as YAML
func RenderYAML(report *DailyReport) ([]byte, error) {
	data, err := yaml.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return data, nil
}

// Render renders the report in the specified format
func Render(report *DailyReport, format string) (string, error) {
	switch format {
	case "plain":
		return RenderPlain(report)
	case "json":
		data, err := RenderJSON(report)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "yaml":
		data, err := RenderYAML(report)
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
