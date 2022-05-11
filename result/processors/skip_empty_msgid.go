package processors

import (
	"time"

	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

type skipEmptyMsgID struct{}

func NewSkipEmptyMsgID() Processor {
	return &skipEmptyMsgID{}
}

func (s skipEmptyMsgID) Process(issues []result.Issue) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "Clean empty msgid")
	return filterIssues(issues, func(i *result.Issue) bool { return i.MsgID != "" }), nil
}

func (s skipEmptyMsgID) Name() string {
	return "skip_empty_msgid"
}

func filterIssues(issues []result.Issue, filter func(i *result.Issue) bool) []result.Issue {
	retIssues := make([]result.Issue, 0, len(issues))
	for i := range issues {
		if filter(&issues[i]) {
			retIssues = append(retIssues, issues[i])
		}
	}

	return retIssues
}
