package processors

import (
	"time"

	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/util"
)

type prepareKey struct{}

func NewPrepareKey() Processor { return &prepareKey{} }

func (p prepareKey) Name() string { return "prepare-key" }

func (p prepareKey) Process(inIssues []extract.Issue) ([]extract.Issue, error) {
	util.TrackTime(time.Now(), "Prepare key")
	outIssues := make([]extract.Issue, 0, len(inIssues))

	for _, iss := range inIssues {
		if iss.IDToken == etype.PluralKey {
			iss.PluralID = iss.MsgID
		}
		outIssues = append(outIssues, iss)
	}
	return outIssues, nil
}
