package visitor

import (
	"go/ast"

	"github.com/iv-menshenin/spider/importwalker/internal/model"
)

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
				v.parent.register(n.Name, o.imported.GetField(s.Sel.String()))
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

func (v *visitor) register(name string, i model.Imported) {
	if i == nil {
		return
	}
	if v.node != nil {
		v.walker.Register(v.node.Pos(), i)
	} else {
		v.walker.Register(0, i)
	}
	if a, ok := v.node.(*ast.AssignStmt); ok {
		for nn, expr := range a.Lhs {
			if ident, ok := expr.(*ast.Ident); ok {
				result := i.GetResult(nn)
				// todo debug
				if result == nil {
					result = i.GetResult(nn)
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

func (a *visitor) makeFileScope(f *ast.File) []obj {
	var scope []obj
	for _, imp := range f.Imports {
		if imp.Path == nil {
			continue
		}
		packagePath := imp.Path.Value[1 : len(imp.Path.Value)-1]
		pkg := a.lookuper.Lookup(packagePath)
		if pkg == nil {
			continue
		}
		var packageTree = pkg.Package("")
		var impAlias = packageTree.Name
		if imp.Name != nil {
			impAlias = imp.Name.String()
		}
		scope = append(scope, obj{
			name:     impAlias,
			imported: importedPackage{path: packagePath, pkg: packageTree, makeFileScope: a.makeFileScope},
		})
	}
	return scope
}

func New(file *ast.File, reg registrar, l model.PackageLookuper) *visitor {
	var v = visitor{
		parent:   nil,
		node:     file,
		walker:   reg,
		lookuper: l,
	}
	v.scope = v.makeFileScope(file)
	return &v
}
