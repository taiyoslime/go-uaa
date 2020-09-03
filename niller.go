package niller

import (
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "niller is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "niller",
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
					switch typ.(type) {
					case *types.Pointer:
						pointerSet[ident.Obj] = []token.Pos{}
					case *types.Slice:
						// TODO
					}
				}
			} else {
				for _, expr := range n.Values {
					switch conexpr := expr.(type) {
					case *ast.CallExpr:
						ident, ok := conexpr.Fun.(*ast.Ident)
						if !ok {
							continue
						}
						decl, ok := ident.Obj.Decl.(*ast.FuncDecl)
						if !ok {
							continue
						}
						var retValType []types.Type
						for _, field:= range decl.Type.Results.List {
							if field.Names == nil {
								typ := pass.TypesInfo.TypeOf(field.Type)
								retValType = append(retValType, typ)
							} else {
								for _, ident := range field.Names {
									typ := pass.TypesInfo.TypeOf(ident)
									retValType = append(retValType, typ)
								}
							}
						}

						for i, typ := range retValType {
							switch typ.(type) {
							case *types.Pointer:
								ident := n.Names[i]
								pointerSet[ident.Obj] = []token.Pos{}
							case *types.Slice:
								// TODO
							}
						}
					}
				}
			}
			return false
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE { // :=
				// TODO
			}
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

