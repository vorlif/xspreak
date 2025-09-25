package extractors

import (
	"context"
	"go/ast"
	"go/types"
	"time"

	"golang.org/x/tools/go/packages"

	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/util"
)

type sliceDefExtractor struct{}

func NewSliceDefExtractor() extract.Extractor {
	return &sliceDefExtractor{}
}

func (v sliceDefExtractor) Run(_ context.Context, extractCtx *extract.Context) ([]extract.Issue, error) {
	util.TrackTime(time.Now(), "extract slices")
	var issues []extract.Issue

	extractCtx.Inspector.Nodes([]ast.Node{&ast.CompositeLit{}}, func(rawNode ast.Node, push bool) (proceed bool) {
		proceed = true
		if !push {
			return
		}

		node := rawNode.(*ast.CompositeLit)
		if len(node.Elts) == 0 {
			return
		}

		arrayTye, ok := node.Type.(*ast.ArrayType)
		if !ok {
			return
		}

		var obj types.Object
		var pkg *packages.Package
		var token etype.Token
		switch val := arrayTye.Elt.(type) {
		case *ast.SelectorExpr:
			pkg, obj = extractCtx.GetType(val.Sel)
			if pkg == nil {
				return
			}
			token = extractCtx.GetLocalizeTypeToken(val.Sel)
		case *ast.Ident:
			pkg, obj = extractCtx.GetType(val)
			if pkg == nil {
				return
			}
			token = extractCtx.GetLocalizeTypeToken(val)
		case *ast.StarExpr:
			switch pointerExpr := val.X.(type) {
			case *ast.SelectorExpr:
				pkg, obj = extractCtx.GetType(pointerExpr.Sel)
				if pkg == nil {
					return
				}
				token = extractCtx.GetLocalizeTypeToken(pointerExpr.Sel)
			case *ast.Ident:
				pkg, obj = extractCtx.GetType(pointerExpr)
				if pkg == nil {
					return
				}
				token = extractCtx.GetLocalizeTypeToken(pointerExpr)

			default:
				return
			}
		default:
			return
		}

		// Array of strings
		if etype.IsMessageID(token) {
			for _, elt := range node.Elts {
				for _, res := range extractCtx.SearchStrings(elt) {
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
			}

			return
		} else if token != etype.None {
			writeMissingMessageID(extractCtx.GetPosition(node.Pos()), token, "")
		}

		structAttr := extractCtx.Definitions.GetFields(util.ObjToKey(obj))
		if structAttr == nil {
			return
		}

		for _, elt := range node.Elts {
			compLit, isCompLit := elt.(*ast.CompositeLit)
			if !isCompLit {
				continue
			}

			structIssues := extractStruct(extractCtx, compLit, obj, pkg)
			issues = append(issues, structIssues...)
		}

		return
	})

	return issues, nil
}

func (v sliceDefExtractor) Name() string {
	return "slicedef_extractor"
}
