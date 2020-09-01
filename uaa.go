package uaa

import (
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "uaa is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "uaa",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func isNil(expr ast.Expr) bool {
	idt, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	return idt.Name == "nil" && idt.Obj == nil
}

func run(pass *analysis.Pass) (interface{}, error) {

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	checkedNilSet := map[*ast.Object]struct{}{}
	pointerSet :=  map[*ast.Object][]token.Pos{}

	inspect.Nodes(nil, func(n ast.Node, push bool) bool{
		switch n := n.(type) {
		case *ast.Ident:
			_, ok := pointerSet[n.Obj]
			if ok {
				pointerSet[n.Obj] = append(pointerSet[n.Obj], n.Pos())
			}
			return false
		case *ast.ValueSpec:
			if n.Values == nil {
				for _, ident := range n.Names {
					typ := pass.TypesInfo.TypeOf(ident)
					if typ != nil {
						_, ok := typ.(*types.Pointer)
						if ok {
							pointerSet[ident.Obj] = []token.Pos{}
						}
					}
				}
			}
			return false

		case *ast.IfStmt:
			switch n1 := n.Cond.(type) {
			case *ast.BinaryExpr:
				if n1.Op.String() == "==" || n1.Op.String() == "!=" {
					if isNil(n1.X) != isNil(n1.Y) {
						var checkedIdent ast.Expr
						if isNil(n1.X) {
							checkedIdent = n1.Y
						} else {
							checkedIdent = n1.X
						}
						obj, ok := checkedIdent.(*ast.Ident)
						if ok {
							checkedNilSet[obj.Obj] = struct{}{}
						}
					}
				}
			}
		}
		return true
	})

	for k := range pointerSet {
		_, ok := checkedNilSet[k]
		if !ok {
			for _, pos := range pointerSet[k] {
				pass.Reportf(pos, "%s may be nil", k.Name)
			}
		}
	}

	return nil, nil
}

