package util

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

func ObjToKey(obj types.Object) string {
	switch v := obj.Type().(type) {
	case *types.Signature:
		if recv := v.Recv(); recv != nil {
			// Strip out the generic type declaration from the type name.
			// The ast.CallExpr reports its receiver as the actual type
			// (e.g.`Generic[string]`), whereas the ast.FuncDecl on the
			// same type as `Generic[T]`. The returned key values need
			// to be consistent between different invocation patterns.
			recv, _, _ := strings.Cut(recv.Type().String(), "[")

			return fmt.Sprintf("%s.%s", recv, obj.Name())
		}

		return fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name())
	case *types.Pointer:
		return v.Elem().String()
	default:
		return fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name())
	}
}

func SearchSelector(expr any) *ast.SelectorExpr {
	current := expr
	for current != nil {
		switch v := current.(type) {
		case *ast.SelectorExpr:
			return v
		case *ast.Ident:
			if v.Obj == nil {
				return nil
			}
			current = v.Obj.Decl
		case *ast.ValueSpec:
			current = v.Type
		case *ast.Field:
			current = v.Type
		case *ast.Ellipsis:
			current = v.Elt
		default:
			return nil
		}
	}
	return nil
}
