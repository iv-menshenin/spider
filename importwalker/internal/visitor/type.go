package visitor

import (
	"go/ast"

	"github.com/iv-menshenin/spider/importwalker/internal/model"
)

type (
	importedType struct {
		fil importedFile
		t   *ast.TypeSpec
	}
)

func (p importedType) GetField(name string) model.Imported {
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

func findIdent(list []*ast.Ident, name string) (*ast.Ident, int) {
	for i, ident := range list {
		if ident.String() == name {
			return ident, i
		}
	}
	return nil, -1
}

func (p importedType) GetResult(int) model.Imported {
	// todo ?
	return nil
}

func (p importedType) GetLevel() []model.Level {
	return []model.Level{model.LevelImportStruct}
}

func (p importedType) GetPackage() string {
	return p.fil.pkg.path
}
