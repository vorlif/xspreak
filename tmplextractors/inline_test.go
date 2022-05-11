package tmplextractors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/extractors"
	"github.com/vorlif/xspreak/goextractors"
)

func TestInlineExtraction(t *testing.T) {
	cfg := config.NewDefault()
	cfg.SourceDir = testdataDir
	cfg.ExtractErrors = false
	require.NoError(t, cfg.Prepare())

	ctx := context.Background()
	contextLoader := extract.NewContextLoader(cfg)

	extractCtx, err := contextLoader.Load(ctx)
	require.NoError(t, err)

	runner, err := extract.NewRunner(cfg, extractCtx.Packages)
	require.NoError(t, err)

	issues, err := runner.Run(ctx, extractCtx, []extractors.Extractor{
		goextractors.NewCommentsExtractor(),
		NewInlineTemplateExtractor(),
		NewCommandExtractor(),
	})
	require.NoError(t, err)

	want := []string{
		"Hello",

		"Dog", "Dogs",

		"Multiline String\nwith\n  newlines",
	}
	got := collectIssueStrings(issues)
	assert.ElementsMatch(t, want, got)

}
