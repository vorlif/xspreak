package processors

import (
	"github.com/vorlif/xspreak/extract"
)

type Processor interface {
	Process(issues []extract.Issue) ([]extract.Issue, error)
	Name() string
}
