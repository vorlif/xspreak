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

type structDefExtractor struct{}

func NewStructDefExtractor() extract.Extractor {
	return &structDefExtractor{}
}

func (v structDefExtractor) Run(_ context.Context, extractCtx *extract.Context) ([]extract.Issue, error) {
	util.TrackTime(time.Now(), "extract structs")
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

		var ident *ast.Ident
		switch val := node.Type.(type) {
		case *ast.SelectorExpr:
			ident = val.Sel
		case *ast.Ident:
			ident = val
		case *ast.IndexExpr:
			switch x := val.X.(type) {
			case *ast.Ident:
				ident = x
			}
		default:
			return
		}

		pkg, obj := extractCtx.GetType(ident)
		if pkg == nil {
			return
		}

		if structAttr := extractCtx.Definitions.GetFields(util.ObjToKey(obj)); structAttr == nil {
			return
		}

		structIssues := extractStruct(extractCtx, node, obj, pkg)
		issues = append(issues, structIssues...)

		return
	})

	return issues, nil
}

func (v structDefExtractor) Name() string {
	return "struct_extractor"
}

func extractStruct(extractCtx *extract.Context, node *ast.CompositeLit, obj types.Object, pkg *packages.Package) []extract.Issue {
	var issues []extract.Issue

	definitionKey := util.ObjToKey(obj)
	collector := newSearchCollector()

	if _, isKv := node.Elts[0].(*ast.KeyValueExpr); isKv {
		for _, elt := range node.Elts {
			kve, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			idt, ok := kve.Key.(*ast.Ident)
			if !ok {
				continue
			}

			def := extractCtx.Definitions.Get(definitionKey, idt.Name)
			if def == nil {
				continue
			}

			foundResults := extractCtx.SearchStrings(kve.Value)
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
	} else {
		for _, attrDef := range extractCtx.Definitions.GetFields(definitionKey) {
			for i, elt := range node.Elts {
				if attrDef.FieldPos != i {
					continue
				}

				foundResults := extractCtx.SearchStrings(elt)
				if len(foundResults) == 0 {
					continue
				}

				switch attrDef.Token {
				case etype.Singular, etype.Key, etype.PluralKey:
					collector.AddSingulars(attrDef.Token, foundResults)
				case etype.Plural:
					collector.Plurals = append(collector.Plurals, foundResults...)
				case etype.Context:
					collector.Contexts = append(collector.Contexts, foundResults...)
				case etype.Domain:
					collector.Domains = append(collector.Domains, foundResults...)
				}
			}
		}
	}

	collector.CheckMissingMessageID(extractCtx)
	for i, singularResult := range collector.Singulars {
		issue := extract.Issue{
			FromExtractor: "extract_struct",
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

	return issues
}
