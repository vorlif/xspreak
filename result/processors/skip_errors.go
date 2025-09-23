package processors

import (
	"slices"
	"time"

	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

type skipErrors struct{}

func NewSkipErrors() Processor { return &skipErrors{} }

func (s skipErrors) Name() string { return "skip-errors" }

func (s skipErrors) Process(issues []result.Issue) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "Skip errors")

	issues = slices.DeleteFunc(issues, func(iss result.Issue) bool { return iss.FromExtractor == "error_extractor" })
	return issues, nil
}
