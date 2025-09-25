package processors

import (
	"slices"
	"time"

	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/util"
)

type skipIgnoreFlag struct{}

// NewSkipIgnore creates a new processor that skips issues with the "ignore" flag.
func NewSkipIgnore() Processor { return &skipIgnoreFlag{} }

func (s skipIgnoreFlag) Name() string { return "skip-ignore" }

func (s skipIgnoreFlag) Process(issues []extract.Issue) ([]extract.Issue, error) {
	util.TrackTime(time.Now(), "Skip ignore")

	issues = slices.DeleteFunc(issues, func(iss extract.Issue) bool {
		return slices.Contains(iss.Flags, "ignore")
	})

	return issues, nil
}
