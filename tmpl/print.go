package tmpl

import (
	"fmt"
	"io"
	"os"
	"text/template/parse"
)

func Fprint(w io.Writer, x *parse.ListNode) error {
	return fprint(w, x)
}

func fprint(w io.Writer, x *parse.ListNode) (err error) {
	// setup printer
	p := printer{
		output: w,
		last:   '\n', // force printing of line number on first line
	}

	// install error handler
	defer func() {
		if e := recover(); e != nil {
			err = e.(localError).err // re-panics if it's not a localError
		}
	}()

	// print x
	if x == nil {
		p.printf("nil\n")
		return
	}
	p.print(x)
	p.printf("\n")

	return
}

// Print prints x to standard output, skipping nil fields.
// Print(fset, x) is the same as Fprint(os.Stdout, fset, x, NotNilFilter).
func Print(x *parse.ListNode) error {
	return Fprint(os.Stdout, x)
}

type printer struct {
	output io.Writer
	indent int  // current indentation level
	last   byte // the last byte processed by Write
	line   int  // current line number
}

var indent = []byte(".  ")

func (p *printer) Write(data []byte) (n int, err error) {
	var m int
	for i, b := range data {
		// invariant: data[0:n] has been written
		if b == '\n' {
			m, err = p.output.Write(data[n : i+1])
			n += m
			if err != nil {
				return
			}
			p.line++
		} else if p.last == '\n' {
			_, err = fmt.Fprintf(p.output, "%6d  ", p.line)
			if err != nil {
				return
			}
			for j := p.indent; j > 0; j-- {
				_, err = p.output.Write(indent)
				if err != nil {
					return
				}
			}
		}
		p.last = b
	}
	if len(data) > n {
		m, err = p.output.Write(data[n:])
		n += m
	}
	return
}

type localError struct {
	err error
}

func (p *printer) printf(format string, args ...any) {
	if _, err := fmt.Fprintf(p, format, args...); err != nil {
		panic(localError{err})
	}
}

func (p *printer) print(node parse.Node) {
	if node == nil {
		p.printf("nil")
	}

	p.printf("\n")

	switch v := node.(type) {
	case *parse.ActionNode:
		p.printf("ActionNode")
		p.indent++
		p.print(v.Pipe)
	case *parse.BoolNode:
		p.printf("BoolNode")
		p.printf("%v", v.True)
	case *parse.BranchNode:
		p.printf("BranchNode")
		p.indent++
		p.print(v.Pipe)
		p.print(v.List)
		p.print(v.ElseList)
	case *parse.BreakNode:
		p.printf("BreakNode %d", v.Line)
		p.indent++
	case *parse.ChainNode:
		p.printf("ChainNode\n")
		p.indent++
		p.print(v.Node)
		for i, str := range v.Field {
			p.printf("Field %d %s\n", i, str)
		}
	case *parse.CommandNode:
		p.printf("CommandNode\n")
		p.indent++
		for _, arg := range v.Args {
			p.print(arg)
		}
	case *parse.CommentNode:
		p.printf("CommentNode")
		p.printf("Text %s", v.Text)
		p.indent++
	case *parse.FieldNode:
		p.printf("FieldNode")
		p.indent++
		for i, str := range v.Ident {
			p.printf("\nIdent %d %s", i, str)
		}
	case *parse.IdentifierNode:
		p.printf("IdentifierNode\n")
		p.indent++
		p.printf("Ident %s", v.Ident)
	case *parse.IfNode:
		p.printf("IfNode")
		p.indent++
		p.print(&v.BranchNode)
	case *parse.ListNode:
		p.printf("ListNode")
		p.indent++
		for _, arg := range v.Nodes {
			p.print(arg)
		}
	case *parse.NumberNode:
		p.printf("NumberNode")
		p.indent++
		p.printf(v.String())
	case *parse.PipeNode:
		p.printf("PipeNode\n")
		p.indent++
		p.printf("IsAssign %v", v.IsAssign)
		for _, arg := range v.Decl {
			p.print(arg)
		}
		for _, arg := range v.Cmds {
			p.print(arg)
		}
	case *parse.RangeNode:
		p.printf("RangeNode")
		p.indent++
		p.print(&v.BranchNode)
	case *parse.StringNode:
		p.printf("StringNode\n")
		p.indent++
		p.printf("Quoted %v\n", v.Quoted)
		p.printf("Text %v", v.Text)
	case *parse.TemplateNode:
		p.printf("TemplateNode")
		p.indent++
		p.printf("Line %v", v.Line)
		p.printf("Name %v", v.Name)
		p.print(v.Pipe)
	case *parse.TextNode:
		p.printf("TextNode\n")
		p.indent++
		p.printf("Text %q", string(v.Text))
	case *parse.VariableNode:
		p.printf("VariableNode")
		p.indent++
		for i, str := range v.Ident {
			p.printf("Ident %d %s", i, str)
		}
	case *parse.WithNode:
		p.printf("WithNode")
		p.indent++
		p.print(&v.BranchNode)
	default:
		p.printf("%T", node)
		p.indent++
	}

	p.indent--
}
