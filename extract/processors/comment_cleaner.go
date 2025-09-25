package processors

import (
	"strings"
	"time"

	"github.com/vorlif/xspreak/config"
	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/util"
)

const flagPrefix = "xspreak:"

type commentCleaner struct {
	allowPrefixes []string
}

var _ Processor = (*commentCleaner)(nil)

func NewCommentCleaner(cfg *config.Config) Processor {
	c := &commentCleaner{
		allowPrefixes: make([]string, 0, len(cfg.CommentPrefixes)),
	}

	for _, prefix := range cfg.CommentPrefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix != "" {
			c.allowPrefixes = append(c.allowPrefixes, prefix)
		}
	}

	return c
}

func (s commentCleaner) Process(inIssues []extract.Issue) ([]extract.Issue, error) {
	util.TrackTime(time.Now(), "Clean comments")
	outIssues := make([]extract.Issue, 0, len(inIssues))

	for _, iss := range inIssues {
		cleanedComments := make([]string, 0)

		// remove duplicates and extract text
		commentLines := make(map[string][]string)
		for _, com := range iss.Comments {
			commentLines[com] = strings.Split(com, "\n")
		}

		// filter text
		for _, lines := range commentLines {
			isTranslatorComment := false

			var cleanedLines []string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if s.hasTranslatorPrefix(line) {
					isTranslatorComment = true
				} else if strings.HasPrefix(line, flagPrefix) {
					iss.Flags = append(iss.Flags, util.ParseFlags(line)...)
					isTranslatorComment = false
					continue
				} else if len(line) == 0 {
					isTranslatorComment = false
					continue
				}

				if isTranslatorComment {
					cleanedLines = append(cleanedLines, line)
				}
			}

			if len(cleanedLines) > 0 {
				cleanedComments = append(cleanedComments, strings.Join(cleanedLines, " "))
			}
		}

		iss.Comments = cleanedComments
		outIssues = append(outIssues, iss)
	}

	return outIssues, nil
}

func (s commentCleaner) hasTranslatorPrefix(line string) bool {
	for _, prefix := range s.allowPrefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	return false
}

func (s commentCleaner) Name() string {
	return "comment_cleaner"
}
