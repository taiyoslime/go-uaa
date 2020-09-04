package niller

import (
	"errors"
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

type obj struct {
	pos         []token.Pos
	errObj      *ast.Object
	provablyNil bool
}

func createObj() *obj {
	return &obj{
		[]token.Pos{}, nil, true,
	}
}

func run(pass *analysis.Pass) (interface{}, error) {

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	objSet := map[*ast.Object]*obj{}
	nilCheckedObjSet := map[*ast.Object]struct{}{}

	errorType := types.Universe.Lookup("error").Type()

	analyzeReturnValOfCallExpr := func(expr *ast.CallExpr) (interface{}, error) {
		// TODO
		ident, ok := expr.Fun.(*ast.Ident)
		if !ok {
			return nil, errors.New("unexpected AST node")
		}
		decl, ok := ident.Obj.Decl.(*ast.FuncDecl)
		if !ok {
			return nil, errors.New("unexpected AST node")
		}
		var retValType []types.Type
		for _, field := range decl.Type.Results.List {
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
		return &retValType, nil
	}

	analyzeCallExpr := func(expr *ast.CallExpr, idents []*ast.Ident) {
		retValType, err := analyzeReturnValOfCallExpr(expr)
		if err != nil {
			return
		}

		var objs []*ast.Object
		var errObjs *ast.Object

		for i, typ := range *(retValType.(*[]types.Type)) {
			ident := idents[i]
			if types.Identical(typ, errorType) {
				errObjs = ident.Obj
			} else {
				switch typ.(type) {
				case *types.Pointer:
					objs = append(objs, ident.Obj)
					objSet[ident.Obj] = createObj()

				case *types.Slice:
					//TODO
				}
			}
		}
		if errObjs == nil {
			for _, obj := range objs {
				objSet[obj].errObj = obj
			}
		} else {
			for _, obj := range objs {
				objSet[obj].errObj = errObjs
			}
		}
	}

	inspect.Nodes(nil, func(n ast.Node, push bool) bool {
		switch n := n.(type) {
		case *ast.Ident:
			obj, ok := objSet[n.Obj]
			if ok && obj.provablyNil {
				objSet[n.Obj].pos = append(objSet[n.Obj].pos, n.Pos())
			}
			return false
		case *ast.ValueSpec:
			if n.Values == nil {
				for _, ident := range n.Names {
					typ := pass.TypesInfo.TypeOf(ident)
					switch typ.(type) {
					case *types.Pointer:
						objSet[ident.Obj] = createObj()
						objSet[ident.Obj].errObj = ident.Obj
					case *types.Slice:
						// TODO
					}
				}
			} else {
				for _, expr := range n.Values {
					switch conexpr := expr.(type) {
					case *ast.CallExpr:
						analyzeCallExpr(conexpr, n.Names)
					}
				}
			}
			return false
		case *ast.AssignStmt: // :=, = etc...
			for _, expr := range n.Lhs {
				ident, ok := expr.(*ast.Ident)
				if !ok {
					continue
				}
				typ := pass.TypesInfo.TypeOf(ident)
				switch typ.(type) {
				case *types.Pointer, *types.Slice:
				default:
					continue
				}
			}

			for i, expr := range n.Rhs {
				switch conexpr := expr.(type) {
				case *ast.CallExpr:
					idents := []*ast.Ident{}
					for _, expr := range n.Lhs {
						ident, ok := expr.(*ast.Ident)
						if !ok {
							continue
						}
						idents = append(idents, ident)
					}
					analyzeCallExpr(conexpr, idents)
				default:
					if n.Tok == token.ASSIGN { // =
						lident, ok := n.Lhs[i].(*ast.Ident)
						if !ok {
							continue
						}
						rident, ok := expr.(*ast.Ident)
						if !ok {
							objSet[lident.Obj].provablyNil = false
							continue
						}
						typ := pass.TypesInfo.TypeOf(rident)
						if types.Identical(typ, types.Typ[types.UntypedNil]) {
							objSet[lident.Obj].provablyNil = true
						} else {
							objSet[lident.Obj].provablyNil = false
						}
					}
				}
			}
			return false
		case *ast.IfStmt:
			switch n1 := n.Cond.(type) {
			case *ast.BinaryExpr:
				if n1.Op == token.EQL || n1.Op == token.NEQ {
					typX := pass.TypesInfo.TypeOf(n1.X)
					typY := pass.TypesInfo.TypeOf(n1.Y)
					if types.Identical(typX, types.Typ[types.UntypedNil]) != types.Identical(typY, types.Typ[types.UntypedNil]) {
						var checkedIdent ast.Expr
						if types.Identical(typX, types.Typ[types.UntypedNil]) {
							checkedIdent = n1.Y
						} else {
							checkedIdent = n1.X
						}
						obj, ok := checkedIdent.(*ast.Ident)
						if ok {
							nilCheckedObjSet[obj.Obj] = struct{}{}
						}
					}
				}
			}
		}
		return true
	})

	for k, v := range objSet {
		errObj := v.errObj
		_, ok := nilCheckedObjSet[errObj]

		if !ok {
			for _, pos := range v.pos {
				pass.Reportf(pos, "%s may be nil", k.Name)
			}
		}
	}

	return nil, nil
}
