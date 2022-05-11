package tmpl

import (
	"text/template/parse"
)

// Based on https://cs.opensource.google/go/x/tools/+/refs/tags/v0.1.10:go/ast/inspector/

const (
	nAction = iota
	nBool
	nBranch
	nBreak
	nChain
	nCommand
	nComment
	nContinue
	nDot
	nField
	nIdentifier
	nIf
	nList
	nNil
	nNumber
	nPipe
	nRange
	nString
	nTemplate
	nText
	NVariable
	nWith
)

type Inspector struct {
	events []event
}

func newInspector(nodes []parse.Node) *Inspector {
	return &Inspector{traverse(nodes)}
}

type event struct {
	node  parse.Node
	typ   uint64 // typeOf(node)
	index int    // 1 + index of corresponding pop event, or 0 if this is a pop
}

func (in *Inspector) Preorder(types []parse.Node, f func(parse.Node)) {
	mask := maskOf(types)
	for i := 0; i < len(in.events); {
		ev := in.events[i]
		if ev.typ&mask != 0 {
			if ev.index > 0 {
				f(ev.node)
			}
		}
		i++
	}
}

func (in *Inspector) Nodes(types []parse.Node, f func(n parse.Node, push bool) (proceed bool)) {
	mask := maskOf(types)
	for i := 0; i < len(in.events); {
		ev := in.events[i]
		if ev.typ&mask != 0 {
			if ev.index > 0 {
				// push
				if !f(ev.node, true) {
					i = ev.index // jump to corresponding pop + 1
					continue
				}
			} else {
				// pop
				f(ev.node, false)
			}
		}
		i++
	}
}

func (in *Inspector) WithStack(types []parse.Node, f func(n parse.Node, push bool, stack []parse.Node) (proceed bool)) {
	mask := maskOf(types)
	var stack []parse.Node
	for i := 0; i < len(in.events); {
		ev := in.events[i]
		if ev.index > 0 {
			// push
			stack = append(stack, ev.node)
			if ev.typ&mask != 0 {
				if !f(ev.node, true, stack) {
					i = ev.index
					stack = stack[:len(stack)-1]
					continue
				}
			}
		} else {
			// pop
			if ev.typ&mask != 0 {
				f(ev.node, false, stack)
			}
			stack = stack[:len(stack)-1]
		}
		i++
	}
}

func traverse(nodes []parse.Node) []event {
	// This estimate is based on the net/http package.
	capacity := len(nodes) * 33 / 100
	if capacity > 1e6 {
		capacity = 1e6 // impose some reasonable maximum
	}
	events := make([]event, 0, capacity)

	var stack []event
	for _, root := range nodes {
		Inspect(root, func(n parse.Node) bool {
			if n != nil {
				// push
				ev := event{
					node:  n,
					typ:   typeOf(n),
					index: len(events), // push event temporarily holds own index
				}
				stack = append(stack, ev)
				events = append(events, ev)
			} else {
				// pop
				ev := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				events[ev.index].index = len(events) + 1 // make push refer to pop

				ev.index = 0 // turn ev into a pop event
				events = append(events, ev)
			}
			return true
		})
	}

	return events
}

func typeOf(n parse.Node) uint64 {
	switch n.(type) {
	case *parse.ActionNode:
		return 1 << nAction
	case *parse.BranchNode:
		return 1 << nBranch
	case *parse.ChainNode:
		return 1 << nChain
	case *parse.CommandNode:
		return 1 << nCommand
	case *parse.IfNode:
		return 1 << nIf
	case *parse.ListNode:
		return 1 << nList
	case *parse.PipeNode:
		return 1 << nPipe
	case *parse.RangeNode:
		return 1 << nRange
	case *parse.TemplateNode:
		return 1 << nTemplate
	case *parse.WithNode:
		return 1 << nWith
	case *parse.BoolNode:
		return 1 << nBool
	case *parse.BreakNode:
		return 1 << nBreak
	case *parse.CommentNode:
		return 1 << nComment
	case *parse.ContinueNode:
		return 1 << nContinue
	case *parse.DotNode:
		return 1 << nDot
	case *parse.FieldNode:
		return 1 << nField
	case *parse.IdentifierNode:
		return 1 << nIdentifier
	case *parse.NilNode:
		return 1 << nNil
	case *parse.NumberNode:
		return 1 << nNumber
	case *parse.StringNode:
		return 1 << nString
	case *parse.TextNode:
		return 1 << nText
	case *parse.VariableNode:
		return 1 << NVariable
	}

	return 0
}

func maskOf(nodes []parse.Node) uint64 {
	if nodes == nil {
		return 1<<64 - 1 // match all node types
	}
	var mask uint64
	for _, n := range nodes {
		mask |= typeOf(n)
	}
	return mask
}
