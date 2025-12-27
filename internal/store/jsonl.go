package store

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/jackchuka/gh-brag/internal/data"
)

// LoadExistingIDs reads the JSONL file and returns a map of existing IDs.
func LoadExistingIDs(filepath string) (map[string]bool, error) {
	existing := make(map[string]bool)
	f, err := os.Open(filepath)
	if os.IsNotExist(err) {
		return existing, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var evt data.Event
		if err := json.Unmarshal(scanner.Bytes(), &evt); err == nil {
			if evt.ID != "" {
				existing[evt.ID] = true
			}
		}
	}
	return existing, scanner.Err()
}

// AppendEvents appends new events to the JSONL file.
func AppendEvents(filepath string, events []data.Event) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	encoder := json.NewEncoder(f)
	for _, evt := range events {
		if err := encoder.Encode(evt); err != nil {
			return err
		}
	}
	return nil
}
