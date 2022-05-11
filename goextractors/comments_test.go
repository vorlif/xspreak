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

func TestCommentsExtractor(t *testing.T) {
	cfg := config.NewDefault()
	cfg.SourceDir = testdataDir
	ctx := context.Background()
	contextLoader := extract.NewContextLoader(cfg)

	extractCtx, err := contextLoader.Load(ctx)
	require.NoError(t, err)

	runner, err := extract.NewRunner(cfg, extractCtx.Packages)
	require.NoError(t, err)

	e := []extractors.Extractor{NewCommentsExtractor()}
	issues, err := runner.Run(ctx, extractCtx, e)
	require.NoError(t, err)
	require.Empty(t, issues)

	assert.NotNil(t, extractCtx.CommentMaps)
	assert.NotEmpty(t, extractCtx.CommentMaps)
}
