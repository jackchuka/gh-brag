package spinner

import (
	"time"

	"github.com/briandowns/spinner"
)

func NewSpinner(suffix string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + suffix
	return s
}
