package goextractors

import (
	"context"
	"go/ast"
	"time"

	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/extract/extractors"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

type globalAssignExtractor struct{}

func NewGlobalAssignExtractor() extractors.Extractor {
	return &globalAssignExtractor{}
}

func (v globalAssignExtractor) Run(_ context.Context, extractCtx *extractors.Context) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "extract global assign")
	var issues []result.Issue

	extractCtx.Inspector.Nodes([]ast.Node{&ast.ValueSpec{}}, func(rawNode ast.Node, push bool) (proceed bool) {
		proceed = true
		if !push {
			return
		}
		node := rawNode.(*ast.ValueSpec)

		selector := searchSelector(node.Type)
		if selector == nil {
			return
		}

		tok := extractCtx.GetLocalizeTypeToken(selector)
		if tok != etype.Singular {
			return
		}

		in, ok := selector.X.(*ast.Ident)
		if !ok {
			return
		}

		pkg, _ := extractCtx.GetType(in)
		if pkg == nil {
			return
		}

		for _, value := range node.Values {
			for _, res := range extractCtx.SearchStrings(value) {
				issue := result.Issue{
					FromExtractor: v.Name(),
					MsgID:         res.Raw,
					Pkg:           pkg,
					Comments:      extractCtx.GetComments(pkg, res.Node),
					Pos:           extractCtx.GetPosition(res.Node.Pos()),
				}

				issues = append(issues, issue)
			}
		}

		return
	})

	return issues, nil
}

func (v globalAssignExtractor) Name() string {
	return "global_assign"
}
