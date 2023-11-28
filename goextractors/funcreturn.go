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

type funcReturnExtractor struct{}

func NewFuncReturnExtractor() extractors.Extractor {
	return &funcReturnExtractor{}
}

func (e *funcReturnExtractor) Run(_ context.Context, extractCtx *extractors.Context) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "extract func return values")
	var issues []result.Issue

	extractCtx.Inspector.WithStack([]ast.Node{&ast.FuncDecl{}}, func(rawNode ast.Node, push bool, stack []ast.Node) (proceed bool) {
		proceed = true
		if !push {
			return
		}

		node := rawNode.(*ast.FuncDecl)
		if node.Body == nil || node.Type == nil || node.Type.Results == nil || len(node.Type.Results.List) == 0 {
			return
		}

		// Extract the return types if from the localise package
		extractedResults := make([]etype.Token, len(node.Type.Results.List))
		var foundType bool
		for i, res := range node.Type.Results.List {
			tok, _ := extractCtx.SearchIdentAndToken(res)
			if tok == etype.None {
				extractedResults[i] = etype.None
				continue
			}

			extractedResults[i] = tok
			foundType = true
		}

		if !foundType {
			return
		}

		pkg, _ := extractCtx.GetType(node.Name)
		if pkg == nil {
			return
		}

		// Extract the values from the return statements
		ast.Inspect(node.Body, func(node ast.Node) bool {
			if node == nil {
				return true
			}

			retNode, isReturn := node.(*ast.ReturnStmt)
			if !isReturn || len(retNode.Results) != len(extractedResults) {
				return true
			}

			collector := newSearchCollector()
			collector.ExtraNodes = append(collector.ExtraNodes, node)

			for i, extractedResult := range extractedResults {
				foundResults := extractCtx.SearchStrings(retNode.Results[i])
				if len(foundResults) == 0 {
					continue
				}

				switch extractedResult {
				case etype.Singular, etype.Key, etype.PluralKey:
					collector.AddSingulars(extractedResult, foundResults)
				case etype.Plural:
					collector.Plurals = append(collector.Plurals, foundResults...)
				case etype.Context:
					collector.Contexts = append(collector.Contexts, foundResults...)
				case etype.Domain:
					collector.Domains = append(collector.Domains, foundResults...)
				}
			}

			collector.CheckMissingMessageID(extractCtx)
			for i, singularResult := range collector.Singulars {
				issue := result.Issue{
					FromExtractor: e.Name(),
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

			return true
		})

		return
	})

	return issues, nil
}

func (e *funcReturnExtractor) Name() string {
	return "func_return"
}
