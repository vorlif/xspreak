package tmplextractors

import (
	"context"
	"strings"
	"text/template/parse"

	log "github.com/sirupsen/logrus"

	"github.com/vorlif/xspreak/extract/extractors"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/tmpl"
)

type commandExtractor struct{}

func NewCommandExtractor() extractors.Extractor {
	return &commandExtractor{}
}

func (c *commandExtractor) Run(_ context.Context, extractCtx *extractors.Context) ([]result.Issue, error) {
	var issues []result.Issue
	if len(extractCtx.Config.Keywords) == 0 {
		log.Debug("Skip template extraction, no keywords present")
		return issues, nil
	}

	for _, template := range extractCtx.Templates {
		template.ExtractComments()

		template.Inspector.WithStack([]parse.Node{&parse.PipeNode{}}, func(n parse.Node, push bool, stack []parse.Node) (proceed bool) {
			proceed = false
			if !push {
				return
			}
			pipe := n.(*parse.PipeNode)
			if pipe.IsAssign {
				return
			}

			for _, cmd := range pipe.Cmds {
				iss := extractIssue(cmd, extractCtx, template)
				if iss == nil {
					continue
				}

				iss.Flags = append(iss.Flags, "go-template")
				issues = append(issues, *iss)
			}

			return
		})
	}
	return issues, nil
}

func (c *commandExtractor) Name() string {
	return "tmpl_command"
}

func extractIssue(cmd *parse.CommandNode, extractCtx *extractors.Context, template *tmpl.Template) *result.Issue {
	if cmd == nil {
		return nil
	}
	raw := cmd.String()
	for _, keyword := range extractCtx.Config.Keywords {
		if !strings.HasPrefix(raw, keyword.Name+" ") {
			continue
		}

		if keyword.MaxPosition() >= len(cmd.Args)-1 { // The first index contains the keyword itself
			log.Warnf("Template keyword found but not enough arguments available: %s %s", template.Position(cmd.Pos), raw)
			continue
		}

		return extractArgs(cmd.Args[1:], keyword, template)
	}

	return nil
}

func extractArgs(args []parse.Node, keyword *tmpl.Keyword, template *tmpl.Template) *result.Issue {
	iss := &result.Issue{}

	if stringNode, ok := args[keyword.SingularPos].(*parse.StringNode); ok {
		iss.MsgID = stringNode.Text
		iss.IDToken = keyword.IDToken
		iss.Comments = template.GetComments(args[keyword.SingularPos].Position())
		iss.Pos = template.Position(args[keyword.SingularPos].Position())
	} else {
		log.Warnf("Template keyword is not passed a string: %s", args[keyword.SingularPos])
		return nil
	}

	if keyword.PluralPos >= 0 {
		if stringNode, ok := args[keyword.PluralPos].(*parse.StringNode); ok {
			iss.PluralID = stringNode.Text
		} else {
			log.Warnf("Template keyword is not passed a string: %s", args[keyword.PluralPos])
			return nil
		}
	}

	if keyword.ContextPos >= 0 {
		if stringNode, ok := args[keyword.ContextPos].(*parse.StringNode); ok {
			iss.Context = stringNode.Text
		} else {
			log.Warnf("Template keyword is not passed a string: %s", args[keyword.ContextPos])
			return nil
		}
	}

	if keyword.DomainPos >= 0 {
		if stringNode, ok := args[keyword.DomainPos].(*parse.StringNode); ok {
			iss.Domain = stringNode.Text
		} else {
			log.Warnf("Template keyword is not passed a string: %s", args[keyword.DomainPos])
			return nil
		}
	}

	return iss
}
