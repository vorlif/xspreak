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

	extractCtx.Inspector.Nodes([]ast.Node{&ast.CallExpr{}}, func(rawNode ast.Node, push bool) (proceed bool) {
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

		// type conversions, e.g. localize.Singular("init")
		if tok := extractCtx.GetLocalizeTypeToken(ident); etype.IsMessageID(tok) {
			for _, res := range extractCtx.SearchStrings(node.Args[0]) {
				issue := result.Issue{
					FromExtractor: v.Name(),
					IDToken:       tok,
					MsgID:         res.Raw,
					Pkg:           pkg,
					Comments:      extractCtx.GetComments(pkg, res.Node),
					Pos:           extractCtx.GetPosition(res.Node.Pos()),
				}

				issues = append(issues, issue)
			}

		}

		funcParameterDefs := extractCtx.Definitions.GetFields(util.ObjToKey(obj))
		if funcParameterDefs == nil {
			return
		}

		collector := newSearchCollector()

		// Function calls
		for _, def := range funcParameterDefs {
			for i, arg := range node.Args {
				if (def.FieldPos != i) && !(i >= def.FieldPos && def.IsVariadic) {
					continue
				}

				foundResults := extractCtx.SearchStrings(arg)
				if len(foundResults) == 0 {
					continue
				}

				switch def.Token {
				case etype.Singular, etype.Key, etype.PluralKey:
					collector.AddSingulars(def.Token, foundResults)
				case etype.Plural:
					collector.Plurals = append(collector.Plurals, foundResults...)
				case etype.Context:
					collector.Contexts = append(collector.Contexts, foundResults...)
				case etype.Domain:
					collector.Domains = append(collector.Domains, foundResults...)
				}
			}
		}

		for i, singularResult := range collector.Singulars {
			issue := result.Issue{
				FromExtractor: v.Name(),
				IDToken:       collector.SingularType[i],
				MsgID:         singularResult.Raw,
				Domain:        collector.GetDomain(),
				Context:       collector.GetContext(),
				PluralID:      collector.GetPlural(),
				Comments:      extractCtx.GetComments(pkg, singularResult.Node),
				Pkg:           pkg,
				Pos:           extractCtx.GetPosition(singularResult.Node.Pos()),
			}
			issues = append(issues, issue)
		}

		return
	})

	return issues, nil
}

func (v funcCallExtractor) Name() string {
	return "funccall_extractor"
}
