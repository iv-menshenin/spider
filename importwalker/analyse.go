package importwalker

import (
	"go/ast"
	"go/token"
	"strings"
)

/*
	todo:
		проснализировать сигнатуры функции, если они не добавляют импортированные структуры (даже своего пакета)
 		то они могут быть представлены в форме интерфейсных

*/

func (w *Walker) startAnalyse() error {
	if len(w.pkgs) == 0 {
		return nil
	}
	for i := range w.pkgs {
		if err := w.analyse(w.pkgs[i].path, w.pkgs[i].pkgs); err != nil {
			return err
		}
	}
	return nil
}

func (w *Walker) analyse(path string, p packages) error {
	for name, pkg := range p {
		if strings.HasSuffix(name, "_test") {
			continue
		}
		for fileName, file := range pkg.Files {
			if strings.HasSuffix(fileName, "_test.go") {
				continue
			}
			w.analyseFile(path, fileName, file)
		}
	}
	return nil
}

type wRegistrar struct {
	p        string
	n        string
	f        *ast.File
	detected []precedent
}

func (r *wRegistrar) register(pos token.Pos, i imported) {
	for _, l := range i.getLevel() {
		r.detected = append(r.detected, precedent{
			dependedOn:  i.getPackage(),
			packageName: r.p,
			fileName:    r.n,
			filePos:     pos,
			level:       l,
		})
	}
}

func (w *Walker) analyseFile(packName, fileName string, f *ast.File) {
	var reg = wRegistrar{p: packName, n: fileName, f: f}
	var v = visitor{
		walker: &reg,
		scope:  w.makeFileScope(f),
		node:   f,
	}
	ast.Walk(&v, f)
	w.precedents = append(w.precedents, reg.detected...)
}

func (w *Walker) makeFileScope(f *ast.File) []obj {
	var scope []obj
	for _, imp := range f.Imports {
		if imp.Path == nil {
			continue
		}
		packagePath := imp.Path.Value[1 : len(imp.Path.Value)-1]
		var nativeName = w.pkgNames[packagePath]
		impAlias := nativeName
		if imp.Name != nil {
			impAlias = imp.Name.String()
		}
		if impAlias == "" {
			continue
		}
		for _, pkg := range w.pkgs {
			if pkg.name == packagePath {
				scope = append(scope, obj{
					name:     impAlias,
					imported: importedPackage{path: packagePath, pkg: pkg.pkgs[nativeName], makeFileScope: w.makeFileScope},
				})
			}
		}
	}
	return scope
}
