package visitor

import (
	"go/ast"
	"go/token"

	"github.com/iv-menshenin/spider/importwalker/internal/model"
)

type (
	importedPackage struct {
		path          string
		makeFileScope func(*ast.File) []obj
		pkg           *ast.Package
	}
	importedFile struct {
		pkg importedPackage
		fil *ast.File
	}
)

func (p importedFile) reference(expr ast.Expr) model.Imported {
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

func (p importedPackage) GetField(name string) model.Imported {
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

func (p importedPackage) reference(expr ast.Expr) model.Imported {
	switch t := expr.(type) {
	case *ast.Ident:
		return p.GetField(t.Name)
	case *ast.StarExpr:
		return p.reference(t.X)
	case *ast.SelectorExpr:
		return nil
	default:
		// todo
		return nil
	}
}

func (p importedPackage) GetResult(nn int) model.Imported {
	if nn > 0 {
		return nil
	}
	return p
}

func (p importedPackage) GetLevel() []model.Level {
	return []model.Level{model.LevelImportNone}
}

func (p importedPackage) GetPackage() string {
	return p.path
}
