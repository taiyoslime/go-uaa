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

func analyzeCallExpr(expr *ast.CallExpr, pass *analysis.Pass) (interface{}, error) {
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

func run(pass *analysis.Pass) (interface{}, error) {

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nilCheckedObjSet := map[*ast.Object]struct{}{}
	objErrMap := map[*ast.Object]*ast.Object{}
	pointerSet := map[*ast.Object][]token.Pos{}

	errorType := types.Universe.Lookup("error").Type()

	inspect.Nodes(nil, func(n ast.Node, push bool) bool {
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
						objErrMap[ident.Obj] = ident.Obj
					case *types.Slice:
						// TODO
					}
				}
			} else {
				for _, expr := range n.Values {
					switch conexpr := expr.(type) {
					case *ast.CallExpr:
						retValType, err := analyzeCallExpr(conexpr, pass)
						if err != nil {
							continue
						}

						var pointerObj []*ast.Object
						var errObj *ast.Object

						for i, typ := range *(retValType.(*[]types.Type)) {
							ident := n.Names[i]
							if types.Identical(typ, errorType) {
								errObj = ident.Obj
							} else {
								switch typ.(type) {
								case *types.Pointer:
									pointerObj = append(pointerObj, ident.Obj)
									pointerSet[ident.Obj] = []token.Pos{}
								case *types.Slice:
									//TODO
								}
							}
						}
						if errObj == nil {
							for _, obj := range pointerObj {
								objErrMap[obj] = obj
							}
						} else {
							for _, obj := range pointerObj {
								objErrMap[obj] = errObj
							}
						}
					}
				}
			}
			return false
		case *ast.AssignStmt: // :=
			for _, expr := range n.Rhs {
				switch conexpr := expr.(type) {
				case *ast.CallExpr:
					retValType, err := analyzeCallExpr(conexpr, pass)
					if err != nil {
						continue
					}

					var pointerObj []*ast.Object
					var errObj *ast.Object

					for i, typ := range *(retValType.(*[]types.Type)) {
						ident := n.Lhs[i].(*ast.Ident)
						if types.Identical(typ, errorType) {
							errObj = ident.Obj
						} else {
							switch typ.(type) {
							case *types.Pointer:
								pointerObj = append(pointerObj, ident.Obj)
								pointerSet[ident.Obj] = []token.Pos{}
							case *types.Slice:
								//TODO
							}
						}
					}
					if errObj == nil {
						for _, obj := range pointerObj {
							objErrMap[obj] = obj
						}
					} else {
						for _, obj := range pointerObj {
							objErrMap[obj] = errObj
						}
					}
				}

			}
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

	for k := range pointerSet {
		errObj := objErrMap[k]
		_, ok := nilCheckedObjSet[errObj]

		if !ok {
			for _, pos := range pointerSet[k] {
				pass.Reportf(pos, "%s may be nil", k.Name)
			}
		}
	}

	return nil, nil
}
