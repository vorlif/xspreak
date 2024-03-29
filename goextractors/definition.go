package goextractors

import (
	"context"
	"go/ast"
	"go/token"
	"strconv"
	"time"

	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract/extractors"
)

type definitionExtractor struct{}

func NewDefinitionExtractor() extractors.Extractor {
	return &definitionExtractor{}
}

func (d *definitionExtractor) Run(ctx context.Context, extractCtx *extractors.Context) ([]result.Issue, error) {
	defer util.TrackTime(time.Now(), "Extract definitions")
	runner := &definitionExtractorRunner{
		ctx:        ctx,
		extractCtx: extractCtx,
	}
	extractCtx.Inspector.Nodes(nil, runner.searchDefinitions)
	return []result.Issue{}, nil
}

func (d definitionExtractor) Name() string {
	return "extract_definitions"
}

type definitionExtractorRunner struct {
	ctx        context.Context
	extractCtx *extractors.Context
}

func (de *definitionExtractorRunner) searchDefinitions(n ast.Node, push bool) bool {
	if !push {
		return true
	}

	switch v := n.(type) {
	case *ast.FuncDecl:
		de.extractFunc(v)
	case *ast.AssignStmt:
		de.extractInlineFunc(v)
	case *ast.GenDecl:
		switch v.Tok {
		case token.VAR:
			de.extractVar(v)
		case token.TYPE:
			de.extractStruct(v)
		}
	}

	return true
}

// var t localize.Singular.
func (de *definitionExtractorRunner) extractVar(decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		selector := searchSelector(valueSpec.Type)
		if selector == nil {
			continue
		}

		tok := de.extractCtx.GetLocalizeTypeToken(selector)
		if tok != etype.Singular {
			// TODO(fv): log hint
			continue
		}

		for _, name := range valueSpec.Names {
			pkg, obj := de.extractCtx.GetType(name)
			if pkg == nil {
				continue
			}

			def := &extractors.Definition{
				Type:  extractors.VarSingular,
				Token: tok,
				Pck:   pkg,
				Ident: name,
				Path:  obj.Pkg().Path(),
				ID:    util.ObjToKey(obj),
				Obj:   obj,
			}

			de.addDefinition(def)
		}

	}

}

/*
	type TT struct {
		T localize.Singular
		P localize.Plural
	}.
*/
func (de *definitionExtractorRunner) extractStruct(decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		pkg, obj := de.extractCtx.GetType(typeSpec.Name)
		if obj == nil {
			continue
		}

		if !config.ShouldScanStruct(pkg.PkgPath) {
			continue
		}

		for i, field := range structType.Fields.List {

			var tok etype.Token
			switch field.Type.(type) {
			case *ast.Ident:
				if pkg.PkgPath == config.SpreakLocalizePackagePath {
					tok = de.extractCtx.GetLocalizeTypeToken(field.Type)
					break
				}
				if selector := searchSelector(field.Type); selector == nil {
					continue
				} else {
					tok = de.extractCtx.GetLocalizeTypeToken(selector)
				}
			default:
				if selector := searchSelector(field.Type); selector == nil {
					continue
				} else {
					tok = de.extractCtx.GetLocalizeTypeToken(selector)
				}
			}

			if tok == etype.None {
				continue
			}

			for ii, fieldName := range field.Names {
				def := &extractors.Definition{
					Type:       extractors.StructField,
					Token:      tok,
					Pck:        pkg,
					Ident:      typeSpec.Name,
					Path:       obj.Pkg().Path(),
					ID:         util.ObjToKey(obj),
					Obj:        obj,
					FieldIdent: fieldName,
					FieldName:  fieldName.Name,
					FieldPos:   calculatePosIdx(ii, i),
				}

				de.addDefinition(def)
			}
		}
	}
}

// func translate(msgid localize.Singular, plural localize.Plural)
// func getTranslation() (localize.Singular, localize.Plural).
func (de *definitionExtractorRunner) extractFunc(decl *ast.FuncDecl) {
	if decl.Type == nil || decl.Type.Params == nil {
		return
	}

	de.extractFunctionsParams(decl.Name, decl.Type)
}

func (de *definitionExtractorRunner) extractInlineFunc(assign *ast.AssignStmt) {
	if len(assign.Lhs) == 0 || len(assign.Lhs) != len(assign.Rhs) {
		return
	}

	ident, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		return
	}

	funcLit, ok := assign.Rhs[0].(*ast.FuncLit)
	if !ok || funcLit.Type == nil || funcLit.Type.Params == nil {
		return
	}

	de.extractFunctionsParams(ident, funcLit.Type)
}

func (de *definitionExtractorRunner) extractFunctionsParams(ident *ast.Ident, t *ast.FuncType) {
	pck, obj := de.extractCtx.GetType(ident)
	if pck == nil {
		return
	}

	// function call
	for i, param := range t.Params.List {
		tok, _ := de.extractCtx.SearchIdentAndToken(param)
		if tok == etype.None {
			continue
		}

		if len(param.Names) == 0 {
			def := &extractors.Definition{
				Type:       extractors.FunctionParam,
				Token:      tok,
				Pck:        pck,
				Ident:      ident,
				Path:       obj.Pkg().Path(),
				ID:         util.ObjToKey(obj),
				Obj:        obj,
				FieldIdent: nil,
				FieldName:  strconv.Itoa(i),

				FieldPos:   i,
				IsVariadic: isEllipsis(param.Type),
			}
			de.addDefinition(def)
		}

		for ii, name := range param.Names {
			def := &extractors.Definition{
				Type:       extractors.FunctionParam,
				Token:      tok,
				Pck:        pck,
				Ident:      ident,
				Path:       obj.Pkg().Path(),
				ID:         util.ObjToKey(obj),
				Obj:        obj,
				FieldIdent: name,
				FieldName:  name.Name,
				IsVariadic: isEllipsis(param.Type),

				FieldPos: calculatePosIdx(i, ii),
			}
			de.addDefinition(def)
		}
	}
}

func (de *definitionExtractorRunner) addDefinition(d *extractors.Definition) {
	key := d.Key()
	if _, ok := de.extractCtx.Definitions[key]; !ok {
		de.extractCtx.Definitions[key] = make(map[string]*extractors.Definition)
	}

	de.extractCtx.Definitions[key][d.FieldName] = d
}

func isEllipsis(node ast.Node) bool {
	switch node.(type) {
	case *ast.Ellipsis:
		return true
	default:
		return false
	}
}
