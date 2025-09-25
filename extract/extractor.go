package extract

import (
	"context"
)

// Extractor uses the context to search for strings to be translated and returns all strings found as an issue.
type Extractor interface {
	Run(ctx context.Context, extractCtx *Context) ([]Issue, error)

	// Name returns the name of the extractor
	Name() string
}
