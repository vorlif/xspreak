package tmplextractors

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/extractors"
	"github.com/vorlif/xspreak/result"
)

var (
	testdataDir       = filepath.FromSlash("../testdata/project")
	testdataTemplates = filepath.FromSlash("../testdata/tmpl")
)

func runExtraction(t *testing.T, dir string, testExtractors ...extractors.Extractor) []result.Issue {
	cfg := config.NewDefault()
	cfg.SourceDir = dir
	cfg.ExtractErrors = false
	cfg.TemplatePatterns = []string{
		testdataTemplates + "/**/*.txt",
		testdataTemplates + "/**/*.html",
		testdataTemplates + "/**/*.tmpl",
	}

	require.NoError(t, cfg.Prepare())

	ctx := context.Background()
	contextLoader := extract.NewContextLoader(cfg)

	extractCtx, err := contextLoader.Load(ctx)
	require.NoError(t, err)

	runner, err := extract.NewRunner(cfg, extractCtx.Packages)
	require.NoError(t, err)

	issues, err := runner.Run(ctx, extractCtx, testExtractors)
	require.NoError(t, err)
	return issues
}

func collectIssueStrings(issues []result.Issue) []string {
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
