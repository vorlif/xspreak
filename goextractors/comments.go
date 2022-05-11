package goextractors

import (
	"context"
	"go/ast"
	"time"

	"github.com/vorlif/xspreak/extract/extractors"
	"github.com/vorlif/xspreak/result"
	"github.com/vorlif/xspreak/util"
)

type commentsExtractor struct{}

func NewCommentsExtractor() extractors.Extractor {
	return &commentsExtractor{}
}

func (d *commentsExtractor) Run(ctx context.Context, extractCtx *extractors.Context) ([]result.Issue, error) {
	util.TrackTime(time.Now(), "Extract comments")
	for _, pkg := range extractCtx.OriginalPackages {
		for _, file := range pkg.Syntax {

			commentMap := ast.NewCommentMap(pkg.Fset, file, file.Comments)
			if len(commentMap) == 0 {
				continue
			}

			if _, hasPkg := extractCtx.CommentMaps[pkg.ID]; !hasPkg {
				extractCtx.CommentMaps[pkg.ID] = make(map[string]ast.CommentMap)
			}

			posit := extractCtx.GetPosition(file.Pos())
			extractCtx.CommentMaps[pkg.ID][posit.Filename] = commentMap
		}
	}

	return []result.Issue{}, nil
}

func (d commentsExtractor) Name() string {
	return "comments_extractor"
}
