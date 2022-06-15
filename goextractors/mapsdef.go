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

type mapsDefExtractor struct{}

func NewMapsDefExtractor() extractors.Extractor {
	return &mapsDefExtractor{}
}

func (v mapsDefExtractor) Run(_ context.Context, extractCtx *extractors.Context) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "extract maps")
	var issues []result.Issue

	extractCtx.Inspector.Nodes([]ast.Node{&ast.CompositeLit{}}, func(rawNode ast.Node, push bool) (proceed bool) {
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

			// Map of strings
			if etype.IsMessageID(token) {
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

					for _, res := range extractCtx.SearchStrings(target) {
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

				continue
			} else if token != etype.None {
				writeMissingMessageID(extractCtx.GetPosition(ident.Pos()), token, "")
			}

			// Array of structs
			structAttr := extractCtx.Definitions.GetFields(util.ObjToKey(obj))
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

				structIssues := extractStruct(extractCtx, compLit, obj, pkg)
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
