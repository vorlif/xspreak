package tmpl

import (
	"fmt"
	"text/template/parse"
)

type Visitor interface {
	Visit(node parse.Node) (w Visitor)
}

func Walk(v Visitor, node parse.Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *parse.ActionNode:
		if n.Pipe != nil {
			Walk(v, n.Pipe)
		}
	case *parse.BranchNode:
		if n.Pipe != nil {
			Walk(v, n.Pipe)
		}
		if n.List != nil {
			Walk(v, n.List)
		}
		if n.ElseList != nil {
			Walk(v, n.ElseList)
		}
	case *parse.ChainNode:
		Walk(v, n.Node)
	case *parse.CommandNode:
		for _, arg := range n.Args {
			Walk(v, arg)
		}
	case *parse.IfNode:
		Walk(v, &n.BranchNode)
	case *parse.ListNode:
		if len(n.Nodes) > 0 {
			for _, arg := range n.Nodes {
				Walk(v, arg)
			}
		}
	case *parse.PipeNode:
		if len(n.Decl) > 0 {
			for _, arg := range n.Decl {
				Walk(v, arg)
			}
		}
		if len(n.Cmds) > 0 {
			for _, arg := range n.Cmds {
				Walk(v, arg)
			}
		}
	case *parse.RangeNode:
		Walk(v, &n.BranchNode)
	case *parse.TemplateNode:
		if n.Pipe != nil {
			Walk(v, n.Pipe)
		}
	case *parse.WithNode:
		Walk(v, &n.BranchNode)
	case *parse.BoolNode, *parse.BreakNode, *parse.CommentNode, *parse.ContinueNode, *parse.DotNode,
		*parse.FieldNode, *parse.IdentifierNode, *parse.NilNode, *parse.NumberNode, *parse.StringNode,
		*parse.TextNode, *parse.VariableNode:
		// Walk(v, n)
	default:
		panic(fmt.Sprintf("tmpl.Walk: unexpected node type %T", n))
	}

	v.Visit(nil)
}

type inspector func(parse.Node) bool

func (f inspector) Visit(node parse.Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}

func Inspect(node parse.Node, f func(parse.Node) bool) {
	Walk(inspector(f), node)
}
