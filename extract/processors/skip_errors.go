package processors

import (
	"slices"
	"time"

	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/util"
)

type skipErrors struct{}

// NewSkipErrors creates a new processor that removes all issues which result from the error extractor.
func NewSkipErrors() Processor { return &skipErrors{} }

func (s skipErrors) Name() string { return "skip-errors" }

func (s skipErrors) Process(issues []extract.Issue) ([]extract.Issue, error) {
	util.TrackTime(time.Now(), "Skip errors")

	issues = slices.DeleteFunc(issues, func(iss extract.Issue) bool { return iss.FromExtractor == "error_extractor" })
	return issues, nil
}
