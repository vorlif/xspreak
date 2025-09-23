package tmpl

import (
	"bufio"
	"bytes"
	"go/token"
	"io"
	"os"
	"strings"
	"text/template/parse"
)

type Template struct {
	Filename  string
	Trees     map[string]*parse.Tree
	Inspector *Inspector

	// GoFilePos holds the position from which .go file the template originates
	GoFilePos token.Position

	// OffsetLookup holds the first position of all line starts.
	OffsetLookup []token.Position

	// Comments holds the comments of each line number
	Comments map[int][]string
}

func Parse(filepath string) (*Template, error) {
	src, errF := os.ReadFile(filepath)
	if errF != nil {
		return nil, errF
	}

	return ParseBytes(filepath, src)
}

func ParseString(name, content string) (*Template, error) {
	return ParseBytes(name, []byte(content))
}

func ParseBytes(name string, src []byte) (*Template, error) {
	t := &Template{
		Filename:  name,
		Trees:     make(map[string]*parse.Tree),
		Comments:  make(map[int][]string),
		Inspector: nil,
	}

	tree := &parse.Tree{
		Name: name,
		Mode: parse.ParseComments | parse.SkipFuncCheck,
	}

	_, err := tree.Parse(string(src), "{{", "}}", t.Trees, map[string]any{})
	if err != nil {
		return nil, err
	}

	roodNotes := make([]parse.Node, 0, len(t.Trees))
	for _, tree := range t.Trees {
		roodNotes = append(roodNotes, tree.Root)
	}

	t.OffsetLookup = extractLineInfos(name, bytes.NewReader(src))
	t.Inspector = newInspector(roodNotes)

	return t, nil
}

func extractLineInfos(filename string, src io.Reader) []token.Position {
	infos := make([]token.Position, 0, 50)
	infos = append(infos, token.Position{Filename: filename, Offset: 0, Line: 1, Column: 1})

	scanner := bufio.NewScanner(src)
	offset := 0
	line := 1
	for scanner.Scan() {
		offset += len(scanner.Text())
		infos = append(infos, token.Position{Filename: filename, Offset: offset, Line: line, Column: len(scanner.Text())})
		offset++
		line++
	}

	return infos
}

func (t *Template) ExtractComments() {
	t.Inspector.Nodes([]parse.Node{&parse.CommentNode{}}, func(rawNode parse.Node, push bool) (proceed bool) {
		proceed = false
		node := rawNode.(*parse.CommentNode)

		comment := strings.TrimSpace(node.Text)
		comment = strings.TrimPrefix(comment, "/*")
		comment = strings.TrimSuffix(comment, "*/")
		comment = strings.TrimSpace(comment)

		pos := t.Position(node.Pos)
		t.Comments[pos.Line] = append(t.Comments[pos.Line], comment)

		return
	})
}

func (t *Template) Position(offset parse.Pos) token.Position {
	var pos token.Position
	pos.Filename = t.Filename

	for i, p := range t.OffsetLookup {
		if p.Offset > int(offset) {
			pos = t.OffsetLookup[i-1]
			break
		}
	}

	if pos.IsValid() {
		pos.Column = int(offset) - pos.Offset
		pos.Offset += int(offset)
	}

	if t.GoFilePos.IsValid() {
		pos.Filename = t.GoFilePos.Filename
		pos.Line += t.GoFilePos.Line - 1
		pos.Column += t.GoFilePos.Column - 1
		pos.Offset += t.GoFilePos.Offset
	}

	return pos
}

func (t *Template) GetComments(offset parse.Pos) []string {
	pos := t.Position(offset)
	var comments []string
	if _, ok := t.Comments[pos.Line]; ok {
		comments = append(comments, t.Comments[pos.Line]...)
	}

	if _, ok := t.Comments[pos.Line-1]; ok {
		comments = append(comments, t.Comments[pos.Line-1]...)
	}

	return comments
}
