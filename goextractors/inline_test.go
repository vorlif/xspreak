package goextractors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/extractors"
)

func TestInlineExtraction(t *testing.T) {
	cfg := config.NewDefault()
	cfg.SourceDir = testdataDir
	cfg.ExtractErrors = true
	require.NoError(t, cfg.Prepare())
	ctx := context.Background()
	contextLoader := extract.NewContextLoader(cfg)

	extractCtx, err := contextLoader.Load(ctx)
	require.NoError(t, err)

	runner, err := extract.NewRunner(cfg, extractCtx.Packages)
	require.NoError(t, err)

	e := []extractors.Extractor{NewCommentsExtractor(), NewInlineTemplateExtractor()}
	issues, err := runner.Run(ctx, extractCtx, e)
	require.NoError(t, err)
	assert.Empty(t, issues)

	assert.Equal(t, 3, len(extractCtx.Templates))
}
