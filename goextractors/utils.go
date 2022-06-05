package goextractors

import (
	"go/ast"

	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/extract/extractors"
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

type searchCollector struct {
	Singulars    []*extractors.SearchResult
	SingularType []etype.Token
	Plurals      []*extractors.SearchResult
	Contexts     []*extractors.SearchResult
	Domains      []*extractors.SearchResult
	ExtraNodes   []ast.Node
}

func newSearchCollector() *searchCollector {
	return &searchCollector{
		Singulars:  make([]*extractors.SearchResult, 0),
		Plurals:    make([]*extractors.SearchResult, 0),
		Contexts:   make([]*extractors.SearchResult, 0),
		Domains:    make([]*extractors.SearchResult, 0),
		ExtraNodes: make([]ast.Node, 0),
	}
}

func (sc *searchCollector) AddSingulars(token etype.Token, singulars []*extractors.SearchResult) {
	sc.Singulars = append(sc.Singulars, singulars...)
	for i := 0; i < len(singulars); i++ {
		sc.SingularType = append(sc.SingularType, token)
	}
}

func (sc *searchCollector) GetPlural() string {
	if len(sc.Plurals) > 0 {
		return sc.Plurals[0].Raw
	}
	return ""
}

func (sc *searchCollector) GetContext() string {
	if len(sc.Contexts) > 0 {
		return sc.Contexts[0].Raw
	}
	return ""
}

func (sc *searchCollector) GetDomain() string {
	if len(sc.Domains) > 0 {
		return sc.Domains[0].Raw
	}
	return ""
}

func (sc *searchCollector) GetNodes() []ast.Node {
	nodes := make([]ast.Node, 0, len(sc.Singulars)+3+len(sc.ExtraNodes))
	nodes = append(nodes, sc.ExtraNodes...)

	for _, sing := range sc.Singulars {
		nodes = append(nodes, sing.Node)
	}
	if len(sc.Plurals) > 0 {
		nodes = append(nodes, sc.Plurals[0].Node)
	}
	if len(sc.Contexts) > 0 {
		nodes = append(nodes, sc.Contexts[0].Node)
	}
	if len(sc.Domains) > 0 {
		nodes = append(nodes, sc.Domains[0].Node)
	}

	return nodes
}
