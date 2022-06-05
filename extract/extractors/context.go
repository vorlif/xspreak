package extractors

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/tmpl"
	"github.com/vorlif/xspreak/util"
)

type Context struct {
	Config *config.Config
	Log    *log.Entry

	OriginalPackages []*packages.Package

	Packages    map[string]*packages.Package
	Inspector   *inspector.Inspector
	CommentMaps map[string]map[string]ast.CommentMap // pkg -> file -> node -> comments

	Definitions Definitions

	Templates []*tmpl.Template
}

func (c *Context) GetPosition(pos token.Pos) token.Position {
	for _, pkg := range c.Packages {
		if position := pkg.Fset.Position(pos); position.IsValid() {
			return position
		}
	}

	return token.Position{}
}

func (c *Context) GetType(ident *ast.Ident) (*packages.Package, types.Object) {
	for _, pkg := range c.Packages {
		if pkg.Types == nil {
			continue
		}
		if obj, ok := pkg.TypesInfo.Defs[ident]; ok {
			if obj == nil || obj.Type() == nil || obj.Pkg() == nil {
				return nil, nil
			}
			return pkg, obj
		}
		if obj, ok := pkg.TypesInfo.Uses[ident]; ok {
			if obj == nil || obj.Type() == nil || obj.Pkg() == nil {
				return nil, nil
			}
			return pkg, obj
		}
		if obj, ok := pkg.TypesInfo.Implicits[ident]; ok {
			if obj == nil || obj.Type() == nil || obj.Pkg() == nil {
				return nil, nil
			}
			return pkg, obj
		}
	}
	return nil, nil
}

func (c *Context) GetLocalizeTypeToken(expr ast.Expr) etype.Token {
	if expr == nil {
		return etype.None
	}

	switch v := expr.(type) {
	case *ast.SelectorExpr:
		return c.GetLocalizeTypeToken(v.Sel)
	case *ast.Ident:
		_, vType := c.GetType(v)
		if vType == nil {
			return etype.None
		}

		if vType.Pkg() == nil || vType.Pkg().Path() != config.SpreakLocalizePackagePath {
			return etype.None
		}

		tok, ok := etype.StringExtractNames[vType.Name()]
		if !ok {
			return etype.None
		}

		return tok
	default:
		return etype.None
	}
}

func (c *Context) SearchIdent(start ast.Node) *ast.Ident {
	switch v := start.(type) {
	case *ast.Ident:
		pkg, _ := c.GetType(v)
		if pkg != nil {
			return v
		}
	case *ast.SelectorExpr:
		pkg, _ := c.GetType(v.Sel)
		if pkg != nil {
			return v.Sel
		}

		return c.SearchIdent(v.X)
	case *ast.StarExpr:
		return c.SearchIdent(v.X)
	}

	return nil
}

func (c *Context) SearchIdentAndToken(start ast.Node) (etype.Token, *ast.Ident) {
	switch val := start.(type) {
	case *ast.Ident:
		if tok := c.GetLocalizeTypeToken(val); tok != etype.None {
			return tok, val
		}

		pkg, obj := c.GetType(val)
		if pkg == nil {
			break
		}

		if def := c.Definitions.Get(util.ObjToKey(obj), ""); def != nil {
			return def.Token, val
		}
	case *ast.StarExpr:
		tok, ident := c.SearchIdentAndToken(val.X)
		if ident != nil {
			pkg, _ := c.GetType(ident)
			if pkg != nil {
				return tok, ident
			}
		}
	}

	selector := searchSelector(start)
	if selector == nil {
		return etype.None, nil
	}

	switch ident := selector.X.(type) {
	case *ast.Ident:
		if tok := c.GetLocalizeTypeToken(ident); tok != etype.None {
			return tok, ident
		}

		pkg, obj := c.GetType(ident)
		if pkg == nil {
			break
		}

		if def := c.Definitions.Get(util.ObjToKey(obj), ""); def != nil {
			return def.Token, ident
		}
		if def := c.Definitions.Get(util.ObjToKey(obj), selector.Sel.Name); def != nil {
			return def.Token, ident
		}

		if obj.Type() == nil {
			break
		}
	}

	if tok := c.GetLocalizeTypeToken(selector.Sel); tok != etype.None {
		return tok, selector.Sel
	}

	pkg, obj := c.GetType(selector.Sel)
	if pkg == nil {
		return etype.None, nil
	}

	if def := c.Definitions.Get(util.ObjToKey(obj), ""); def != nil {
		return def.Token, selector.Sel
	}
	if def := c.Definitions.Get(util.ObjToKey(obj), selector.Sel.Name); def != nil {
		return def.Token, selector.Sel
	}

	return etype.None, nil
}

type SearchResult struct {
	Raw  string
	Node ast.Node
}

func (c *Context) SearchStrings(startExpr ast.Expr) []*SearchResult {
	results := make([]*SearchResult, 0)

	// String was created at the current position
	extracted, originNode := ExtractStringLiteral(startExpr)
	if extracted != "" {
		results = append(results, &SearchResult{Raw: extracted, Node: originNode})
	}

	// Backtracking the string
	startIdent, ok := startExpr.(*ast.Ident)
	if !ok || startIdent.Obj == nil {
		return results
	}
	mainMessage := extracted
	c.Inspector.WithStack([]ast.Node{&ast.AssignStmt{}}, func(raw ast.Node, push bool, stack []ast.Node) (proceed bool) {
		proceed = false

		node := raw.(*ast.AssignStmt)
		if len(node.Lhs) != len(node.Rhs) || len(node.Lhs) == 0 {
			return
		}
		for i, left := range node.Lhs {
			leftIdent, isIdent := left.(*ast.Ident)
			if !isIdent {
				continue
			}
			if leftIdent.Obj != startIdent.Obj {
				continue
			}
			extracted, originNode = ExtractStringLiteral(node.Rhs[i])
			if extracted != "" && extracted != mainMessage {
				results = append(results, &SearchResult{Raw: extracted, Node: node})
			}

		}
		return
	})

	return results
}

// GetComments extracts the Go comments for a list of nodes.
func (c *Context) GetComments(pkg *packages.Package, nodes ...ast.Node) []string {
	var comments []string
	found := make(map[*ast.CommentGroup]bool)
	for _, node := range nodes {
		c.Inspector.WithStack([]ast.Node{node}, func(n ast.Node, push bool, stack []ast.Node) (proceed bool) {
			proceed = false
			// Search stack for our node
			if n != node {
				return
			}

			if _, hasPkg := c.CommentMaps[pkg.PkgPath]; !hasPkg {
				return
			}

			// Find the first node of the line
			pos := c.GetPosition(node.Pos())
			if _, hasFile := c.CommentMaps[pkg.PkgPath][pos.Filename]; !hasFile {
				return
			}

			var topNode = node
			for i := len(stack) - 1; i >= 0; i-- {
				entry := stack[i]
				entryPos := c.GetPosition(entry.Pos())
				if !entryPos.IsValid() || entryPos.Line < pos.Line {
					break
				}

				topNode = entry
			}

			// Search for all comments for this line
			ast.Inspect(topNode, func(node ast.Node) bool {
				nodeComments := c.CommentMaps[pkg.PkgPath][pos.Filename][node]
				for _, com := range nodeComments {
					if _, known := found[com]; known {
						continue
					}

					found[com] = true
					comments = append(comments, com.Text())
				}
				return true
			})
			return
		})
	}

	return comments
}

func (c *Context) BuildPackages() {
	c.Packages = make(map[string]*packages.Package, len(c.OriginalPackages))
	defer util.TrackTime(time.Now(), "Collect packages")
	for _, originalPackage := range c.OriginalPackages {
		c.collectPackages(originalPackage, 0)
	}

	files := make([]*ast.File, 0, 200)
	for _, pkg := range c.Packages {
		files = append(files, pkg.Syntax...)
	}

	c.Inspector = inspector.New(files)
}

func (c *Context) collectPackages(startPck *packages.Package, depth int) {
	if c.Config.MaxDepth >= 0 && depth >= c.Config.MaxDepth {
		return
	}

	c.Packages[startPck.ID] = startPck
	for _, importedPackage := range startPck.Imports {
		if !c.isPartOfDirectory(importedPackage) {
			continue
		}

		if _, ok := c.Packages[importedPackage.ID]; ok {
			continue
		}
		c.collectPackages(importedPackage, depth+1)
	}
}

func (c *Context) isPartOfDirectory(pkg *packages.Package) bool {
	if config.IsValidSpreakPackage(pkg.PkgPath) {
		return true
	}

	for _, src := range c.OriginalPackages {
		if strings.HasPrefix(pkg.PkgPath, src.PkgPath) {
			return true
		}
	}

	return false
}

func searchSelector(expr interface{}) *ast.SelectorExpr {
	switch v := expr.(type) {
	case *ast.SelectorExpr:
		return v
	case *ast.Ident:
		if v.Obj == nil {
			break
		}
		return searchSelector(v.Obj.Decl)
	case *ast.ValueSpec:
		return searchSelector(v.Type)
	case *ast.Field:
		return searchSelector(v.Type)
	case *ast.Ellipsis:
		return searchSelector(v.Elt)
	}
	return nil
}

func ExtractStringLiteral(expr ast.Expr) (string, ast.Node) {
	stack := []ast.Expr{expr}
	var b strings.Builder
	var elem ast.Expr

	for len(stack) != 0 {
		n := len(stack) - 1
		elem = stack[n]
		stack = stack[:n]

		switch v := elem.(type) {
		//  Simple string with quotes or backqoutes
		case *ast.BasicLit:
			if v.Kind != token.STRING {
				continue
			}

			if unqouted, err := strconv.Unquote(v.Value); err != nil {
				b.WriteString(v.Value)
			} else {
				b.WriteString(unqouted)
			}
		// Concatenation of several string literals
		case *ast.BinaryExpr:
			if v.Op != token.ADD {
				continue
			}
			stack = append(stack, v.Y, v.X)
		case *ast.Ident:
			if v.Obj == nil {
				continue
			}
			switch z := v.Obj.Decl.(type) {
			case *ast.ValueSpec:
				if len(z.Values) == 0 {
					continue
				}
				stack = append(stack, z.Values[0])
			case *ast.AssignStmt:
				if len(z.Rhs) == 0 {
					continue
				}
				stack = append(stack, z.Rhs...)
			}
		default:
			continue
		}
	}

	return b.String(), elem
}
