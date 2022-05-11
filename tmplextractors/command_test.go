package tmplextractors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/extractors"
	"github.com/vorlif/xspreak/tmpl"
)

func TestVariablesExtractorRun(t *testing.T) {
	issues := runExtraction(t, testdataDir, NewCommandExtractor())
	assert.NotEmpty(t, issues)

	want := []string{
		"Get-Singular",
		"Getf-Singular",

		"NGet-Singular", "NGet-Plural",
		"NGetf-Singular", "NGetf-Plural",

		"DGet-Domain", "DGet-Singular",
		"DGetf-Domain", "DGetf-Singular",

		"DNGet-Domain", "DNGet-Singular", "DNGet-Plural",
		"DNGetf-Domain", "DNGetf-Singular", "DNGetf-Plural",

		"PGet-Context", "PGet-Singular",
		"PGetf-Context", "PGetf-Singular",

		"DPGet-Domain", "DPGet-Context", "DPGet-Singular",
		"DPGetf-Domain", "DPGetf-Context", "DPGetf-Singular",

		"NPGet-Context", "NPGet-Singular", "NPGet-Plural",
		"NPGetf-Context", "NPGetf-Singular", "NPGetf-Plural",

		"DNPGet-Context", "DNPGet-Context", "DNPGet-Singular", "DNPGet-Plural",
		"DNPGetf-Context", "DNPGetf-Context", "DNPGetf-Singular", "DNPGetf-Plural",
	}
	got := collectIssueStrings(issues)
	assert.ElementsMatch(t, want, got)
}

func TestKeyword(t *testing.T) {
	cfg := config.NewDefault()
	cfg.SourceDir = testdataDir
	cfg.ExtractErrors = false
	cfg.Keywords = []*tmpl.Keyword{
		{
			Name:        ".i18n.Tr",
			SingularPos: 0,
			PluralPos:   -1,
			ContextPos:  -1,
			DomainPos:   -1,
		},
		{
			Name:        ".i18n.Trp",
			SingularPos: 1,
			PluralPos:   3,
			ContextPos:  -1,
			DomainPos:   -1,
		},
	}
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

	issues, err := runner.Run(ctx, extractCtx, []extractors.Extractor{NewCommandExtractor()})
	require.NoError(t, err)

	want := []string{
		"custom keyword",

		"trp-singular", "trp-plural",
	}
	got := collectIssueStrings(issues)
	assert.ElementsMatch(t, want, got)
}
