package extract

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"

	"github.com/vorlif/xspreak/extract/etype"
)

type DefinitionType int

const (
	VarSingular DefinitionType = iota
	Array
	FunctionReturn
	FunctionParam
	StructField
)

// Definition represents a data type in the program code that uses values from the spreak/localize package.
//
// Is used to determine that a text must be extracted when using the data type.
//
// Example:
//
//	func noop(sing alias.MsgID, plural alias.Plural) {}
//	type M struct { Test localize.Singular}
//	var applicationName alias.Singular
type Definition struct {
	Type  DefinitionType
	Token etype.Token
	Pck   *packages.Package
	Ident *ast.Ident
	Path  string // github.com/name/repo/package/pack
	ID    string // github.com/name/repo/package/pack.StructName
	Obj   types.Object

	// -- BEGIN: Only for functions and structs --
	FieldIdent *ast.Ident

	// FieldName is the name of the function parameter or the name of the struct field.
	// Example:
	//  For the fuction noop(sing alias.MsgID, plu alias.Plural) the FieldName is "sing" or "plu".
	//  For the struct M { Test localize.Singular} the FieldName is "Test".
	FieldName  string
	IsVariadic bool
	// -- END: Only for functions and structs --

	// FieldPos is the position of the parameter within the function definition.
	FieldPos int
}

func (d *Definition) Key() string {
	return d.ID
}

// Definitions Is a map of all definitions used in the source code.
//
// path.name -> field || "" -> Definition
// Example:
//
//	github.com/vorlif/testdata.noop -> sing -> Definition
//	github.com/vorlif/testdata.noop -> plural -> Definition
type Definitions map[string]map[string]*Definition

func (defs Definitions) Get(key, fieldName string) *Definition {
	if _, ok := defs[key]; !ok {
		return nil
	}

	if _, ok := defs[key][fieldName]; !ok {
		return nil
	}

	return defs[key][fieldName]
}

func (defs Definitions) GetFields(key string) map[string]*Definition {
	if _, ok := defs[key]; !ok {
		return nil
	}

	if len(defs[key]) == 0 {
		return nil
	}

	return defs[key]
}
