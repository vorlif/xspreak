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

type funcCallExtractor struct{}

func NewFuncCallExtractor() extractors.Extractor {
	return &funcCallExtractor{}
}

func (v funcCallExtractor) Run(_ context.Context, extractCtx *extractors.Context) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "extract func calls")
	var issues []result.Issue

	extractCtx.Inspector.WithStack([]ast.Node{&ast.CallExpr{}}, func(rawNode ast.Node, push bool, stack []ast.Node) (proceed bool) {
		proceed = true
		if !push {
			return
		}

		node := rawNode.(*ast.CallExpr)
		if len(node.Args) == 0 {
			return
		}

		var ident *ast.Ident
		switch fun := node.Fun.(type) {
		case *ast.Ident:
			ident = fun
		case *ast.IndexExpr:
			switch x := fun.X.(type) {
			case *ast.Ident:
				ident = x
			}
		}

		if ident == nil {
			if selector := searchSelector(node.Fun); selector != nil {
				ident = selector.Sel
			} else {
				return
			}
		}

		pkg, obj := extractCtx.GetType(ident)
		if pkg == nil {
			return
		}

		if tok := extractCtx.GetLocalizeTypeToken(ident); etype.IsMessageID(tok) {
			raw, stringNode := extractors.ExtractStringLiteral(node.Args[0])
			if raw == "" {
				return
			}

			issue := result.Issue{
				FromExtractor: v.Name(),
				IDToken:       tok,
				MsgID:         raw,
				Pkg:           pkg,
				Comments:      extractCtx.GetComments(pkg, stringNode, stack),
				Pos:           extractCtx.GetPosition(stringNode.Pos()),
			}

			issues = append(issues, issue)
		}

		funcParameterDefs := extractCtx.Definitions.GetFields(util.ObjToKey(obj))
		if funcParameterDefs == nil {
			return
		}

		issue := result.Issue{
			FromExtractor: v.Name(),
			Pkg:           pkg,
			Pos:           extractCtx.GetPosition(node.Args[0].Pos()),
			Comments:      extractCtx.GetComments(pkg, node.Args[0], stack),
		}
		for _, def := range funcParameterDefs {
			for i, arg := range node.Args {
				if (def.FieldPos != i) && !(i >= def.FieldPos && def.IsVariadic) {
					continue
				}

				for _, res := range extractCtx.SearchStrings(arg) {
					switch def.Token {
					case etype.Singular, etype.Key, etype.PluralKey:
						if issue.MsgID != "" {
							issues = append(issues, issue)
							issue = result.Issue{
								FromExtractor: v.Name(),
								Pkg:           pkg,
								Comments:      extractCtx.GetComments(pkg, node.Args[0], stack),
							}
						}

						issue.Pos = extractCtx.GetPosition(res.Node.Pos())
						issue.IDToken = def.Token
						issue.MsgID = res.Raw
					case etype.Plural:
						issue.PluralID = res.Raw
					case etype.Context:
						issue.Context = res.Raw
					case etype.Domain:
						issue.Domain = res.Raw
					}
				}
			}
		}

		issues = append(issues, issue)
		return
	})

	return issues, nil
}

func (v funcCallExtractor) Name() string {
	return "funccall_extractor"
}
