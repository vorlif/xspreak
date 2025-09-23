package processors

import (
	"slices"
	"time"

	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

type skipEmptyMsgID struct{}

// NewSkipEmptyMsgID create a new processor that removes issues with an empty msgId.
func NewSkipEmptyMsgID() Processor { return &skipEmptyMsgID{} }

func (s skipEmptyMsgID) Name() string { return "skip_empty_msgid" }

func (s skipEmptyMsgID) Process(issues []result.Issue) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "Clean empty msgid")
	issues = slices.DeleteFunc(issues, func(iss result.Issue) bool { return iss.MsgID == "" })
	return issues, nil
}
