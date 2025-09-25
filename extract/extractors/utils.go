package extractors

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/etype"
)

var workingDir, _ = os.Getwd()

type searchCollector struct {
	Singulars    []*extract.SearchResult
	SingularType []etype.Token
	Plurals      []*extract.SearchResult
	Contexts     []*extract.SearchResult
	Domains      []*extract.SearchResult
	ExtraNodes   []ast.Node
}

func writeMissingMessageID(position token.Position, token etype.Token, text string) {
	var typeName string
	switch token {
	case etype.Plural:
		typeName = "plural"
	case etype.Context:
		typeName = "context"
	case etype.Domain:
		typeName = "domain"
	default:
		typeName = "unknown"
	}

	filename := position.Filename
	if relPath, err := filepath.Rel(workingDir, filename); err == nil {
		filename = relPath
	}

	if text != "" {
		log.Warnf("%s:%d usage of %s without MessageID is not supported", filename, position.Line, typeName)
	} else {
		log.Warnf("%s:%d usage of %s without MessageID is not supported: %q", filename, position.Line, typeName, text)
	}
}

func newSearchCollector() *searchCollector {
	return &searchCollector{
		Singulars:  make([]*extract.SearchResult, 0),
		Plurals:    make([]*extract.SearchResult, 0),
		Contexts:   make([]*extract.SearchResult, 0),
		Domains:    make([]*extract.SearchResult, 0),
		ExtraNodes: make([]ast.Node, 0),
	}
}

func (sc *searchCollector) AddSingulars(token etype.Token, singulars []*extract.SearchResult) {
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
		for _, c := range sc.Contexts {
			if c != nil && c.Raw != "" {
				return c.Raw
			}
		}
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

func (sc *searchCollector) CheckMissingMessageID(extractCtx *extract.Context) {
	for _, sing := range sc.Singulars {
		if sing.Raw != "" {
			return
		}
	}

	for _, plural := range sc.Plurals {
		writeMissingMessageID(extractCtx.GetPosition(plural.Node.Pos()), etype.Plural, plural.Raw)
	}

	for _, ctx := range sc.Contexts {
		writeMissingMessageID(extractCtx.GetPosition(ctx.Node.Pos()), etype.Plural, ctx.Raw)
	}

	for _, domain := range sc.Domains {
		writeMissingMessageID(extractCtx.GetPosition(domain.Node.Pos()), etype.Plural, domain.Raw)
	}
}
