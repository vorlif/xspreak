package goextractors

import (
	"context"
	"go/ast"
	"go/token"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/tools/go/packages"

	"github.com/vorlif/xspreak/extract/extractors"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/tmpl"
	"github.com/vorlif/xspreak/util"
)

type inlineTemplateExtractor struct{}

func NewInlineTemplateExtractor() extractors.Extractor {
	return &inlineTemplateExtractor{}
}

func (i *inlineTemplateExtractor) Run(_ context.Context, extractCtx *extractors.Context) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "extract inline templates")

	if len(extractCtx.Config.Keywords) == 0 {
		log.Debug("Skip inline template extraction, no keywords present")
		return []result.Issue{}, nil
	}

	extractCtx.Inspector.WithStack([]ast.Node{&ast.BasicLit{}}, func(rawNode ast.Node, push bool, stack []ast.Node) (proceed bool) {
		proceed = true
		if !push {
			return
		}

		node := rawNode.(*ast.BasicLit)
		if node.Kind != token.STRING {
			return
		}
		// Search for ident to get the package
		var pkg *packages.Package
		for i := len(stack) - 1; i >= 0; i-- {
			if stack[i] == nil {
				break
			}
			if ident := extractIdent(stack[i]); ident != nil {
				pkg, _ = extractCtx.GetType(ident)
				if pkg != nil {
					break
				}
			}
		}

		if pkg == nil {
			return
		}

		comments := extractCtx.GetComments(pkg, node, stack)
		if comments == nil {
			return
		}
		pos := extractCtx.GetPosition(node.Pos())
		templateString, err := strconv.Unquote(node.Value)
		if err != nil {
			return
		}

		for _, comment := range comments {
			if !util.IsInlineTemplate(comment) {
				continue
			}

			template, errP := tmpl.ParseString(pos.Filename, templateString)
			if errP != nil {
				log.WithError(errP).WithField("pos", pos).Warn("Template could not be parsed")
				break
			}
			template.GoFilePos = pos
			extractCtx.Templates = append(extractCtx.Templates, template)
		}

		return
	})

	return []result.Issue{}, nil
}

func (i *inlineTemplateExtractor) Name() string {
	return "inline_template_extractor"
}

func extractIdent(node ast.Node) *ast.Ident {
	switch v := node.(type) {
	case *ast.Ident:
		return v
	case *ast.ValueSpec:
		if len(v.Names) > 0 {
			return v.Names[0]
		}
	case *ast.SelectorExpr:
		return v.Sel
	case *ast.CallExpr:
		if ident, ok := v.Fun.(*ast.Ident); ok {
			return ident
		}
	case *ast.StarExpr:
		switch pointerExpr := v.X.(type) {
		case *ast.SelectorExpr:
			return pointerExpr.Sel
		}
	case *ast.KeyValueExpr:
		if ident, ok := v.Key.(*ast.Ident); ok {
			return ident
		}
		if ident, ok := v.Value.(*ast.Ident); ok {
			return ident
		}
	}

	return nil
}
