package importwalker

import (
	"go/ast"
	"go/token"
)

type (
	registrar interface {
		register(token.Pos, imported)
	}
	imported interface {
		getField(string) imported
		getResult(int) imported
		getLevel() []level
		getPackage() string
	}
	obj struct {
		name     string
		imported imported
	}
	visitor struct {
		parent *visitor
		node   ast.Node
		walker registrar

		scope []obj
	}
	importedPackage struct {
		path          string
		makeFileScope func(*ast.File) []obj
		pkg           *ast.Package
	}
	importedFile struct {
		pkg importedPackage
		fil *ast.File
	}
	importedFunc struct {
		fil importedFile
		fun *ast.FuncDecl
	}
	importedType struct {
		fil importedFile
		t   *ast.TypeSpec
	}
)

func (p importedPackage) getField(name string) imported {
	for fi, f := range p.pkg.Files {
		var fil = importedFile{
			pkg: p,
			fil: p.pkg.Files[fi],
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					if d.Name.String() == name {
						return importedFunc{fil: fil, fun: d}
					}
				}
			case *ast.GenDecl:
				switch d.Tok {
				case token.TYPE:
					for _, spec := range d.Specs {
						if t, ok := spec.(*ast.TypeSpec); ok {
							if t.Name.String() == name {
								return importedType{fil: fil, t: t}
							}
						}
					}
				}
				// todo: constants, variables, types
			}
		}
	}
	return nil
}

func (p importedPackage) reference(expr ast.Expr) imported {
	switch t := expr.(type) {
	case *ast.Ident:
		return p.getField(t.Name)
	case *ast.StarExpr:
		return p.reference(t.X)
	case *ast.SelectorExpr:
		return nil
	default:
		// todo
		return nil
	}
}

func (p importedPackage) getResult(nn int) imported {
	if nn > 0 {
		return nil
	}
	return p
}

func (p importedPackage) getLevel() []level {
	return []level{levelImportNone}
}

func (p importedPackage) getPackage() string {
	return p.path
}

func (p importedFile) reference(expr ast.Expr) imported {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return p.reference(t.X)
	case *ast.SelectorExpr:
		for _, imp := range p.fil.Imports {
			// todo look at the visitor scope
			if imp.Name.String() == "" { // t.X {
				// todo
				println("todo")
			}
		}
		return nil
	default:
		return p.pkg.reference(expr)
	}
}

func (p importedFunc) getField(string) imported {
	return nil
}

func (p importedFunc) getResult(nn int) imported {
	if p.fun.Type.Results == nil {
		return nil
	}
	var i = 0
	for _, ret := range p.fun.Type.Results.List {
		if len(ret.Names) == 0 {
			if i == nn {
				return p.fil.reference(ret.Type)
			}
			i++
		}
		for range ret.Names {
			if i == nn {
				return p.fil.reference(ret.Type)
			}
			i++
		}
	}
	return nil
}

func (p importedFunc) getLevel() []level {
	var result []level
	if p.fun.Recv != nil {
		result = append(result, levelImportMethod)
	} else {
		result = append(result, levelImportFunc)
	}
	// todo: check returned types and add levelImportStruct
	return result
}

func (p importedFunc) getPackage() string {
	return p.fil.pkg.path
}

func (p importedType) getField(name string) imported {
	// todo
	if s, ok := p.t.Type.(*ast.StructType); ok {
		for _, f := range s.Fields.List {
			if _, nn := findIdent(f.Names, name); nn > -1 {
				return p.fil.reference(f.Type)
			}
		}
	}
	for fi, f := range p.fil.pkg.pkg.Files {
		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv != nil {
				if fn.Name.String() != name {
					continue
				}
				for _, recv := range fn.Recv.List {
					var rt = recv.Type
					if se, ok := rt.(*ast.StarExpr); ok {
						rt = se.X
					}
					var recvName string
					switch t := rt.(type) {
					case *ast.Ident:
						recvName = t.String()
					default:
						recvName = "_"
					}
					if recvName == p.t.Name.String() {
						return importedFunc{
							fil: importedFile{
								pkg: p.fil.pkg,
								fil: p.fil.pkg.pkg.Files[fi],
							},
							fun: fn,
						}
					}
				}
			}
		}
	}
	return nil
}

func (p importedType) getResult(int) imported {
	// todo ?
	return nil
}

func (p importedType) getLevel() []level {
	return []level{levelImportStruct}
}

func (p importedType) getPackage() string {
	return p.fil.pkg.path
}

func (v *visitor) Visit(node ast.Node) (w ast.Visitor) {
	if _, ok := node.(*ast.StarExpr); ok {
		return v
	}
	next := v.inherit(node)
	switch n := node.(type) {
	case *ast.Ident:
		s, isSelector := v.node.(*ast.SelectorExpr)
		if isSelector {
			if o, ok := v.selectObj(n.Name); ok {
				v.parent.register(n.Name, o.imported.getField(s.Sel.String()))
				// v.walker.findDeclaration(o.source, o.name, s.Sel.String())
				// v.detectedSelector(n.Name, s.Sel.String())
			}
		}
		// parent.node == CallExpr - MethodCall (dont forget about call-arguments)
		//              AssignStmt - Import ? struct | method | function | interface
		//               RangeStmt - ImportVar ( := range some.VarName )

		//              BinaryExpr - lib.Expr == lib2.Expr2   ImportVar | Function

		// I need to ignore ast.StarExpression - pass-through
		// ast.StarExpression -> ast.TypeAssertExpr ( err.(*lib.ErrType) )
	}
	return &next
}

func (v *visitor) inherit(node ast.Node) visitor {
	return visitor{
		parent: v,
		node:   node,
		walker: v.walker,
		scope:  append([]obj{}, v.scope...),
	}
}

func (v *visitor) selectObj(name string) (result obj, ok bool) {
	for i, o := range v.scope {
		if o.name == name {
			result = v.scope[i]
			ok = true
		}
	}
	return
}

func (v *visitor) register(name string, i imported) {
	if i == nil {
		return
	}
	if v.node != nil {
		v.walker.register(v.node.Pos(), i)
	} else {
		v.walker.register(0, i)
	}
	if a, ok := v.node.(*ast.AssignStmt); ok {
		for nn, expr := range a.Lhs {
			if ident, ok := expr.(*ast.Ident); ok {
				result := i.getResult(nn)
				// todo debug
				if result == nil {
					result = i.getResult(nn)
				}
				if v.parent != nil && ident.String() != "_" && result != nil {
					v.parent.scope = append(v.parent.scope, obj{
						name:     ident.String(),
						imported: result,
					})
				}
			}
		}
		return
	}
	if v.parent != nil {
		v.parent.register(name, i)
	}
}

func findIdent(list []*ast.Ident, name string) (*ast.Ident, int) {
	for i, ident := range list {
		if ident.String() == name {
			return ident, i
		}
	}
	return nil, -1
}
