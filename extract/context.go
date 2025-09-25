package extract

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/tmpl"
	"github.com/vorlif/xspreak/util"
)

// Context holds all the important information needed to scan the program code and template files.
type Context struct {
	Config *config.Config
	Log    *log.Entry

	// OriginalPackages contains the packages that were loaded by golang.org/x/tools/go/packages.
	// It contains the packages of the scanned directory and may contain duplicates.
	OriginalPackages []*packages.Package

	// Packages contains the packages that are of interest to us
	//
	// In addition include:
	// * Packages of the scanned directory
	// * Packages of the Spreak library
	Packages []*packages.Package

	// Inspector to run through the AST of all Packages.
	Inspector *inspector.Inspector

	// CommentMaps for quick access to comments.
	CommentMaps Comments

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

	selector := util.SearchSelector(start)
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
	visited := make(map[ast.Node]bool)

	// String was created at the current position
	extracted, originNode := StringLiteral(startExpr)
	if extracted != "" {
		results = append(results, &SearchResult{Raw: extracted, Node: originNode})
		visited[originNode] = true
	}

	// Backtracking the string
	startIdent, ok := startExpr.(*ast.Ident)
	if !ok {
		return results
	}

	if startIdent.Obj == nil {
		_, obj := c.GetType(startIdent)
		if constObj, ok := obj.(*types.Const); ok && constObj.Val().Kind() == constant.String {
			if stringVal := constant.StringVal(constObj.Val()); stringVal != "" {
				results = append(results, &SearchResult{Raw: stringVal, Node: originNode})
			}
		}
		return results
	}

	c.Inspector.WithStack([]ast.Node{&ast.AssignStmt{}}, func(raw ast.Node, _ bool, _ []ast.Node) (proceed bool) {
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

			if visited[node.Rhs[i]] {
				continue
			}

			extracted, originNode = StringLiteral(node.Rhs[i])
			if extracted != "" {
				visited[node.Rhs[i]] = true
				results = append(results, &SearchResult{Raw: extracted, Node: originNode})
			}

		}
		return
	})

	return results
}

// GetComments extracts the Go comments for a list of nodes.
func (c *Context) GetComments(pkg *packages.Package, node ast.Node) []string {
	var comments []string

	pkgComments, pkgHashComments := c.CommentMaps[pkg.PkgPath]
	if !pkgHashComments {
		return comments
	}

	pos := c.GetPosition(node.Pos())

	fileComments, fileHasComments := pkgComments[pos.Filename]
	if !fileHasComments {
		return comments
	}

	visited := make(map[*ast.CommentGroup]bool)

	c.Inspector.WithStack([]ast.Node{node}, func(n ast.Node, _ bool, stack []ast.Node) (proceed bool) {
		proceed = false
		// Search stack for our node
		if n != node {
			return
		}

		// Find the first node of the line
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
			nodeComments := fileComments[node]
			for _, comment := range nodeComments {
				if visited[comment] {
					continue
				}

				visited[comment] = true
				comments = append(comments, comment.Text())
			}
			return true
		})
		return
	})

	return comments
}

// StringLiteral extracts and concatenates string literals from an AST expression.
// It returns the extracted string and the originating AST node.
func StringLiteral(expr ast.Expr) (string, ast.Node) {
	stack := []ast.Expr{expr}
	var b strings.Builder
	var elem ast.Expr

	for len(stack) != 0 {
		n := len(stack) - 1
		elem = stack[n]
		stack = stack[:n]

		switch v := elem.(type) {
		//  Simple string with quotes or backquotes
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
