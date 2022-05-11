package processors

import (
	"github.com/vorlif/xspreak/result"
)

type Processor interface {
	Process(issues []result.Issue) ([]result.Issue, error)
	Name() string
}
