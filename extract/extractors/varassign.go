package extractors

import (
	"context"
	"go/ast"
	"go/types"
	"time"

	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/util"
)

type varAssignExtractor struct{}

func NewVariablesExtractor() extract.Extractor {
	return &varAssignExtractor{}
}

func (v varAssignExtractor) Run(_ context.Context, extractCtx *extract.Context) ([]extract.Issue, error) {
	util.TrackTime(time.Now(), "extract var assign")
	var issues []extract.Issue

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

		pkg, obj := extractCtx.GetType(ident)
		if pkg == nil {
			return
		}

		if etype.IsMessageID(token) {
			for _, res := range extractCtx.SearchStrings(node.Rhs[0]) {
				issue := extract.Issue{
					FromExtractor: v.Name(),
					IDToken:       token,
					MsgID:         res.Raw,
					Pkg:           pkg,
					Comments:      extractCtx.GetComments(pkg, res.Node),
					Pos:           extractCtx.GetPosition(res.Node.Pos()),
				}

				issues = append(issues, issue)
			}
		} else if token != etype.None {
			shouldPrint := true
			objType := obj.Type()
			if objType != nil {
				if pointerT, ok := objType.(*types.Pointer); ok {
					objType = pointerT.Elem()
				}

				if _, isNamed := objType.(*types.Named); isNamed {
					shouldPrint = false
				}
			}

			if shouldPrint {
				writeMissingMessageID(extractCtx.GetPosition(node.Pos()), token, "")
			}
		}

		return
	})

	return issues, nil
}

func (v varAssignExtractor) Name() string {
	return "varassign_extractor"
}
