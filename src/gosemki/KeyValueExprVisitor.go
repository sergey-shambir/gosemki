package main

import (
	"go/ast"
)

type KeyValueExprVisitor struct {
	indexer *PackageIndexer
	funName *ast.Ident
}

func (this *KeyValueExprVisitor) InspectNode(node ast.Node) bool {
	switch x := node.(type) {
	case *ast.Ident:
		this.funName = x
		return false
	case *ast.CallExpr:
		ast.Inspect(x, this.indexer.InspectNode)
		return false
	case *ast.KeyValueExpr:
		ast.Inspect(x, this.indexer.InspectNode)
		return false
	}
	return true
}

func (this *KeyValueExprVisitor) ApplyIdent() {
	if this.funName != nil {
		pos := this.indexer.fset.Position(this.funName.NamePos)
		goRange := GoRange{
			GoPos: GoPos{
				Line:   pos.Line,
				Column: pos.Column,
				Offset: pos.Offset,
			},
			Length: len(this.funName.Name),
			Kind:   GoKindField,
		}
		this.indexer.result.AddRange(goRange)
	}
}
