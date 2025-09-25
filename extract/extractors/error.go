package extractors

import (
	"context"
	"go/ast"
	"time"

	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/util"

	"github.com/vorlif/xspreak/config"
)

type errorExtractor struct{}

func NewErrorExtractor() extract.Extractor {
	return &errorExtractor{}
}

func (v errorExtractor) Run(_ context.Context, extractCtx *extract.Context) ([]extract.Issue, error) {
	util.TrackTime(time.Now(), "extract errors")
	var issues []extract.Issue

	extractCtx.Inspector.Nodes([]ast.Node{&ast.CallExpr{}}, func(rawNode ast.Node, push bool) (proceed bool) {
		proceed = true
		if !push {
			return
		}

		node := rawNode.(*ast.CallExpr)
		if len(node.Args) != 1 {
			return
		}

		selector := util.SearchSelector(node.Fun)
		if selector == nil {
			return
		}

		pkg, obj := extractCtx.GetType(selector.Sel)
		if pkg == nil {
			return
		}

		if obj.Pkg().Path() != "errors" || !config.ShouldExtractPackage(pkg.PkgPath) {
			return
		}

		msgID, msgNode := extract.StringLiteral(node.Args[0])
		if msgID == "" {
			return
		}

		issue := extract.Issue{
			FromExtractor: v.Name(),
			MsgID:         msgID,
			Pkg:           pkg,
			Context:       extractCtx.Config.ErrorContext,
			Comments:      extractCtx.GetComments(pkg, msgNode),
			Pos:           extractCtx.GetPosition(msgNode.Pos()),
		}

		issues = append(issues, issue)

		return
	})

	return issues, nil
}

func (v errorExtractor) Name() string {
	return "error_extractor"
}
