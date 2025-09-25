package extractors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/loader"
	"github.com/vorlif/xspreak/extract/runner"
	"github.com/vorlif/xspreak/tmpl"
)

func TestInlineExtraction(t *testing.T) {
	cfg := config.NewDefault()
	cfg.SourceDir = testdataDir
	cfg.ExtractErrors = true
	cfg.Keywords = tmpl.DefaultKeywords("T", false)
	require.NoError(t, cfg.Prepare())
	ctx := context.Background()
	contextLoader := loader.NewPackageLoader(cfg)

	extractCtx, err := contextLoader.Load(ctx)
	require.NoError(t, err)

	runner, err := runner.New(cfg, extractCtx.Packages)
	require.NoError(t, err)

	e := []extract.Extractor{NewInlineTemplateExtractor()}
	issues, err := runner.Run(ctx, extractCtx, e)
	require.NoError(t, err)
	assert.Empty(t, issues)

	assert.Equal(t, 3, len(extractCtx.Templates))
}
