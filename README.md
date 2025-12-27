# 🚀 gh-brag

**gh-brag** is a GitHub CLI extension designed to help you aggregate your engineering impact for performance reviews, brag docs, or personal retrospectives. It scans your GitHub activity (PRs, Issues, Reviews), analyzes it for high-value insights, and presents it in a Terminal UI (TUI) dashboard.

<img src="assets/screenshot.png" alt="gh-brag Dashboard Preview" width="300">

---

## ✨ Features

- **🎯 Impact Scoring**: Automatically calculates an "Impact Score" based on weighted actions and thematic focus.
- **📊 TUI Dashboard**: A stunning terminal interface showing:
  - **Theme Distribution**: See where you're spending your time (Feature, Refactor, Bugfix, etc.).
  - **Activity Intensity**: A sleek heatmap of your contributions over time.
  - **Top Repositories**: Identify where you've had the most significant presence.
- **🤝 Collaboration Network**: Visualize your "Review Council" (who reviews you) and your "Mentorship Impact" (who you review).
- **🔧 Customizable**: Flexible theme and metric configuration.

---

## 🔍 Theme Matching

`gh-brag` theme matching order:

1. **Labels First (Priority)**: Searches the PR/Issue's labels that contain theme keywords.
2. **Title Fallback (First Appearance)**: Searches the PR/Issue title for theme keywords. The theme with the keyword appearing **earliest** (lowest index) in the title wins.

---

## 📦 Installation

Install as a `gh` extension:

```bash
gh extension install jackchuka/gh-brag
```

---

## 🛠 Usage

Get insights into your GitHub activity in two simple steps:

### 1. Collect your activity

Gather your GitHub events for a specific period. By default, it looks at the last 6 months.

```bash
gh brag collect
```

_Creates `gh-brag.events.jsonl` containing your raw activity._

### 2. Visualize your impact

Launch the TUI dashboard to explore your insights.

```bash
gh brag visualize
```

---

## 🔍 Advanced Usage

### Customizing the Period

```bash
gh brag collect --from 2024-06-01 --to 2024-12-31
```

### Exporting a YAML Report

If you need a raw data report for your records:

```bash
gh brag analyze --out my-report.yaml
```

---

## ⚙️ Configuration

Customize the analysis by providing a `config.yaml` file. You can adjust theme keyword mappings and impact weights.

```yaml
themes:
  - name: "Feature"
    keywords: ["feat", "feature", "new", "implement"]
  - name: "Refactor"
    keywords: ["refactor", "cleanup", "refactor"]

metrics:
  ownership_threshold: 5 # Min PRs to be considered an 'Owner'
  action_weights:
    merged: 10.0
    reviewed: 5.0
    authored: 2.0
```

Run with your config:

```bash
gh brag --config my-config.yaml visualize
```

---

## 🧰 Tech Stack

- **Go**: Core logic and performance.
- **Cobra**: CLI framework.
- **Lipgloss / Charmbracelet**: Beautiful terminal UI components.
- **GitHub API (via go-gh)**: Reliable data collection.

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.
