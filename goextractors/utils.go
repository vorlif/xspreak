package goextractors

import (
	"go/ast"
)

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
	}
	return nil
}

func calculatePosIdx(first, second int) int {
	if first > 0 && second > 0 {
		return first * second
	} else if first > 0 {
		return first
	} else {
		return second
	}
}
