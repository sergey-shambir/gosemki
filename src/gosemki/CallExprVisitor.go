package main

import (
    "go/ast"
)

type CallExprVisitor struct {
    indexer *PackageIndexer
    funName *ast.Ident
    otherNodes []ast.Node
}

func (this *CallExprVisitor) InspectNode(node ast.Node) bool {
    switch x := node.(type) {
    case *ast.Ident:
        oldNode := this.funName
        if oldNode != nil && oldNode.Obj != nil && oldNode.Obj.Kind != ast.Bad {
            this.indexer.AddIdentRange(this.funName)
        }
        this.funName = x
        return false
    case *ast.KeyValueExpr:
        ast.Inspect(x, this.indexer.InspectNode)
        return false
    case *ast.CallExpr:
        ast.Inspect(x, this.indexer.InspectNode)
        return false
    }
    return true
}

func (this *CallExprVisitor) ApplyIdent() {
    for _, v := range this.otherNodes {
        ast.Inspect(v, this.indexer.InspectNode)
    }
    if this.funName != nil {
        pos := this.indexer.fset.Position(this.funName.NamePos)
        goRange := GoRange {
            GoPos: GoPos {
                Line: pos.Line,
                Column: pos.Column,
                Offset: pos.Offset,
            },
            Length: len(this.funName.Name),
            Kind: GoKindFunc,
        }
        this.indexer.result.AddRange(goRange)
    }
}
