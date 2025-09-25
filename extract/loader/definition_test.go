package loader

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/config"
)

func TestDefinitionExtractor(t *testing.T) {
	cfg := config.NewDefault()
	cfg.SourceDir = testdataDir
	require.NoError(t, cfg.Prepare())
	ctx := context.Background()
	contextLoader := NewPackageLoader(cfg)

	extractCtx, err := contextLoader.Load(ctx)
	require.NoError(t, err)

	defs := extractCtx.Definitions

	key := "github.com/vorlif/testdata.M"
	if assert.Contains(t, defs, key) {
		assert.Contains(t, defs[key], "Test")
	}

	key = "github.com/vorlif/testdata.methodStruct.Method"
	if assert.Contains(t, defs, key) {
		assert.Contains(t, defs[key], "0")
	}

	key = "github.com/vorlif/testdata.genericMethodStruct.Method"
	if assert.Contains(t, defs, key) {
		assert.Contains(t, defs[key], "0")
	}

	key = "github.com/vorlif/testdata.noop"
	if assert.Contains(t, defs, key) {
		assert.Contains(t, defs[key], "sing")
		assert.Contains(t, defs[key], "plural")
		assert.Contains(t, defs[key], "context")
		assert.Contains(t, defs[key], "domain")
	}

	key = "github.com/vorlif/testdata.multiNamesFunc"
	if assert.Contains(t, defs, key) {
		assert.Contains(t, defs[key], "a")
		assert.Contains(t, defs[key], "b")
	}

	key = "github.com/vorlif/testdata.noParamNames"
	if assert.Contains(t, defs, key) {
		assert.Contains(t, defs[key], "0")
		assert.Contains(t, defs[key], "1")
	}

	key = "github.com/vorlif/testdata.variadicFunc"
	if assert.Contains(t, defs, key) {
		if assert.Contains(t, defs[key], "a") {
			assert.Equal(t, 0, defs[key]["a"].FieldPos)
			assert.False(t, defs[key]["a"].IsVariadic)
		}

		if assert.Contains(t, defs[key], "vars") {
			assert.Equal(t, 1, defs[key]["vars"].FieldPos)
			assert.True(t, defs[key]["vars"].IsVariadic)
		}
	}
}
