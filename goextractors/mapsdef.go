package goextractors

import (
	"context"
	"go/ast"
	"time"

	"github.com/vorlif/xspreak/extract/extractors"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

type mapsDefExtractor struct{}

func NewMapsDefExtractor() extractors.Extractor {
	return &mapsDefExtractor{}
}

func (v mapsDefExtractor) Run(_ context.Context, extractCtx *extractors.Context) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "extract maps")
	var issues []result.Issue

	extractCtx.Inspector.WithStack([]ast.Node{&ast.CompositeLit{}}, func(rawNode ast.Node, push bool, stack []ast.Node) (proceed bool) {
		proceed = true
		if !push {
			return
		}

		node := rawNode.(*ast.CompositeLit)
		if len(node.Elts) == 0 {
			return
		}

		mapT, isMap := node.Type.(*ast.MapType)
		if !isMap {
			return
		}

		for _, expr := range []ast.Expr{mapT.Key, mapT.Value} {

			ident := extractCtx.SearchIdent(expr)
			if ident == nil {
				continue
			}

			pkg, obj := extractCtx.GetType(ident)
			if pkg == nil {
				continue
			}

			token := extractCtx.GetLocalizeTypeToken(ident)

			// Array of strings
			if token == extractors.TypeSingular {
				for _, elt := range node.Elts {
					kvExpr, isKv := elt.(*ast.KeyValueExpr)
					if !isKv {
						continue
					}
					var target ast.Expr
					if expr == mapT.Key {
						target = kvExpr.Key
					} else {
						target = kvExpr.Value
					}

					msgID, stringNode := ExtractStringLiteral(target)
					if msgID == "" {
						continue
					}

					issue := result.Issue{
						FromExtractor: v.Name(),
						MsgID:         msgID,
						Pkg:           pkg,
						Comments:      extractCtx.GetComments(pkg, stringNode, stack),
						Pos:           extractCtx.GetPosition(stringNode.Pos()),
					}

					issues = append(issues, issue)
				}

				continue
			}

			structAttr := extractCtx.Definitions.GetFields(objToKey(obj))
			if structAttr == nil {
				return
			}

			for _, elt := range node.Elts {
				kvExpr, isKv := elt.(*ast.KeyValueExpr)
				if !isKv {
					continue
				}
				var target ast.Expr
				if expr == mapT.Key {
					target = kvExpr.Key
				} else {
					target = kvExpr.Value
				}

				compLit, isCompLit := target.(*ast.CompositeLit)
				if !isCompLit {
					continue
				}

				structIssues := extractStruct(extractCtx, compLit, obj, pkg, stack)
				issues = append(issues, structIssues...)
			}
		}

		return
	})

	return issues, nil
}

func (v mapsDefExtractor) Name() string {
	return "mapsdef_extractor"
}
