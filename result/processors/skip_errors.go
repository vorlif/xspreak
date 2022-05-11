package processors

import (
	"time"

	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

type skipErrors struct{}

func NewSkipErrors() Processor {
	return &skipErrors{}
}

func (s skipErrors) Process(issues []result.Issue) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "Skip errors")
	return filterIssues(issues, func(i *result.Issue) bool { return i.FromExtractor != "error_extractor" }), nil
}

func (s skipErrors) Name() string {
	return "skip_errors"
}
