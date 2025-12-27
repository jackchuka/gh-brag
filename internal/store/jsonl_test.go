package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jackchuka/gh-brag/internal/data"
)

func TestJSONL(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.jsonl")

	t.Run("LoadExistingIDs - File does not exist", func(t *testing.T) {
		t.Parallel()

		ids, err := LoadExistingIDs(filepath.Join(tmpDir, "nonexistent.jsonl"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ids) != 0 {
			t.Errorf("expected 0 IDs, got %d", len(ids))
		}
	})

	t.Run("Append and Load", func(t *testing.T) {
		t.Parallel()

		events := []data.Event{
			{ID: "event-1", Title: "Event 1"},
			{ID: "event-2", Title: "Event 2"},
		}

		if err := AppendEvents(testPath, events); err != nil {
			t.Fatalf("failed to append events: %v", err)
		}

		// Verify file exists and has content
		_, err := os.Stat(testPath)
		if err != nil {
			t.Fatalf("expected file to exist: %v", err)
		}

		ids, err := LoadExistingIDs(testPath)
		if err != nil {
			t.Fatalf("failed to load IDs: %v", err)
		}

		if len(ids) != 2 {
			t.Errorf("expected 2 IDs, got %d", len(ids))
		}
		if !ids["event-1"] || !ids["event-2"] {
			t.Errorf("missing expected IDs in %v", ids)
		}
	})

	t.Run("Append to existing file", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(tmpDir, "append.jsonl")

		e1 := []data.Event{{ID: "1"}}
		if err := AppendEvents(path, e1); err != nil {
			t.Fatal(err)
		}

		e2 := []data.Event{{ID: "2"}}
		if err := AppendEvents(path, e2); err != nil {
			t.Fatal(err)
		}

		ids, err := LoadExistingIDs(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) != 2 {
			t.Errorf("expected 2 IDs after appends, got %d", len(ids))
		}
	})
}
