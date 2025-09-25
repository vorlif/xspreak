package extract

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/packages"

	"github.com/vorlif/xspreak/extract/etype"
)

// Issue represents a single issue found by an extractor.
type Issue struct {
	// FromExtractor is the name of the extractor that found this issue.
	FromExtractor string

	IDToken etype.Token

	Domain   string
	Context  string
	MsgID    string
	PluralID string

	Comments []string
	Flags    []string

	Pkg *packages.Package

	Pos token.Position
}

func (i *Issue) FilePath() string    { return i.Pos.Filename }
func (i *Issue) Line() int           { return i.Pos.Line }
func (i *Issue) Column() int         { return i.Pos.Column }
func (i *Issue) Description() string { return fmt.Sprintf("%s: %s", i.FromExtractor, i.MsgID) }
