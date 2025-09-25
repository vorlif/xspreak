package extractors

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/loader"
	"github.com/vorlif/xspreak/extract/runner"
)

const testdataDir = "../../testdata/project"

func TestPrintAst(t *testing.T) {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, filepath.Join(testdataDir, "funccall.go"), nil, 0)
	require.NoError(t, err)

	err = ast.Print(fset, f)
	assert.NoError(t, err)
}

func runExtraction(t *testing.T, dir string, testExtractors ...extract.Extractor) []extract.Issue {
	cfg := config.NewDefault()
	cfg.SourceDir = dir
	cfg.ExtractErrors = true
	require.NoError(t, cfg.Prepare())
	ctx := context.Background()
	contextLoader := loader.NewPackageLoader(cfg)

	extractCtx, err := contextLoader.Load(ctx)
	require.NoError(t, err)

	runner, err := runner.New(cfg, extractCtx.Packages)
	require.NoError(t, err)

	var e []extract.Extractor
	if len(testExtractors) > 0 {
		e = append(e, testExtractors...)
	}
	issues, err := runner.Run(ctx, extractCtx, e)
	require.NoError(t, err)
	return issues
}

func collectIssueStrings(issues []extract.Issue) []string {
	collection := make([]string, 0, len(issues))
	for _, issue := range issues {
		collection = append(collection, issue.MsgID)
		if issue.PluralID != "" {
			collection = append(collection, issue.PluralID)
		}

		if issue.Context != "" {
			collection = append(collection, issue.Context)
		}

		if issue.Domain != "" {
			collection = append(collection, issue.Domain)
		}
	}
	return collection
}
