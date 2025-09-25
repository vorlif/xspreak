package loader

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/config"
)

const testdataDir = "../../testdata/project"

func TestCommentsExtractor(t *testing.T) {
	cfg := config.NewDefault()
	cfg.SourceDir = testdataDir
	ctx := context.Background()

	contextLoader := NewPackageLoader(cfg)

	extractCtx, err := contextLoader.Load(ctx)
	require.NoError(t, err)

	assert.NotNil(t, extractCtx.CommentMaps)
	assert.NotEmpty(t, extractCtx.CommentMaps)
}
