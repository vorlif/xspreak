package loader

import (
	"go/ast"
	"time"

	"golang.org/x/tools/go/packages"

	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/util"
)

// extractComments extracts all comments from the transferred packages and processes them in such a way
// that they can be easily accessed.
func extractComments(pkgs []*packages.Package) extract.Comments {
	util.TrackTime(time.Now(), "Extract comments")

	comments := make(extract.Comments)

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			position := pkg.Fset.Position(file.Pos())
			if !position.IsValid() {
				continue
			}

			commentMap := ast.NewCommentMap(pkg.Fset, file, file.Comments)
			if len(commentMap) == 0 {
				continue
			}

			if _, hasPkg := comments[pkg.ID]; !hasPkg {
				comments[pkg.ID] = make(map[string]ast.CommentMap)
			}
			comments[pkg.ID][position.Filename] = commentMap
		}
	}

	return comments
}
