package extract

import "go/ast"

type Comments map[string]map[string]ast.CommentMap // pkg -> file -> node -> comments
