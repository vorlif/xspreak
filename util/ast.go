package util

import (
	"fmt"
	"go/types"
)

func ObjToKey(obj types.Object) string {
	switch v := obj.Type().(type) {
	case *types.Signature:
		if recv := v.Recv(); recv != nil {
			return fmt.Sprintf("%s.%s", recv, obj.Name())
		}

		return fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name())
	case *types.Pointer:
		return v.Elem().String()
	default:
		return fmt.Sprintf("%s.%s", obj.Pkg().Path(), obj.Name())
	}
}
