package processors

import (
	"time"

	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

type prepareKey struct{}

func NewPrepareKey() Processor { return &prepareKey{} }

func (p prepareKey) Name() string { return "prepare-key" }

func (p prepareKey) Process(inIssues []result.Issue) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "Prepare key")
	outIssues := make([]result.Issue, 0, len(inIssues))

	for _, iss := range inIssues {
		if iss.IDToken == etype.PluralKey {
			iss.PluralID = iss.MsgID
		}
		outIssues = append(outIssues, iss)
	}
	return outIssues, nil
}
