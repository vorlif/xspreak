package extractors

import (
	"context"

	"github.com/vorlif/xspreak/result"
)

type Extractor interface {
	Run(ctx context.Context, extractCtx *Context) ([]result.Issue, error)
	Name() string
}
