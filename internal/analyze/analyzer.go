package analyze

import (
	"errors"

	"github.com/jackchuka/gh-brag/internal/config"
)

// Analyzer encapsulates the analysis configuration and provides methods for various analyses.
type Analyzer struct {
	config *config.Config
}

// New creates a new Analyzer instance with the provided configuration.
func New(config *config.Config) (*Analyzer, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	return &Analyzer{
		config: config,
	}, nil
}
