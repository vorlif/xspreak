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

type varAssignExtractor struct{}

func NewVariablesExtractor() extractors.Extractor {
	return &varAssignExtractor{}
}

func (v varAssignExtractor) Run(_ context.Context, extractCtx *extractors.Context) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "extract var assign")
	var issues []result.Issue

	extractCtx.Inspector.Nodes([]ast.Node{&ast.AssignStmt{}}, func(rawNode ast.Node, push bool) (proceed bool) {
		proceed = true
		if !push {
			return
		}

		node := rawNode.(*ast.AssignStmt)
		if len(node.Lhs) == 0 || len(node.Rhs) == 0 {
			return
		}

		token, ident := extractCtx.SearchIdentAndToken(node.Lhs[0])
		if token == etype.None {
			return
		}

		pkg, _ := extractCtx.GetType(ident)
		if pkg == nil {
			return
		}

		if etype.IsMessageID(token) {
			for _, res := range extractCtx.SearchStrings(node.Rhs[0]) {
				issue := result.Issue{
					FromExtractor: v.Name(),
					IDToken:       token,
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

func (v varAssignExtractor) Name() string {
	return "varassign_extractor"
}
