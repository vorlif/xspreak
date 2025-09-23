package processors

import (
	"slices"
	"time"

	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

type skipIgnoreFlag struct{}

// NewSkipIgnore creates a new processor that skips issues with the "ignore" flag.
func NewSkipIgnore() Processor { return &skipIgnoreFlag{} }

func (s skipIgnoreFlag) Name() string { return "skip-ignore" }

func (s skipIgnoreFlag) Process(issues []result.Issue) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "Skip ignore")

	issues = slices.DeleteFunc(issues, func(iss result.Issue) bool {
		for _, flag := range iss.Flags {
			if flag == "ignore" {
				return true
			}
		}

		return false
	})

	return issues, nil
}
