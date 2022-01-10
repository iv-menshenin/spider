package importwalker

import (
	"github.com/iv-menshenin/spider/importwalker/internal/model"
	"go/ast"
	"go/token"
	"strings"
)

type (
	analyser struct {
		tree       model.AST
		precedents []precedent
	}
)

func (w *Walker) startAnalyse(tree model.AST) error {
	if tree.Len() == 0 {
		return nil
	}
	var a = analyser{
		tree:       tree,
		precedents: []precedent{},
	}
	for p := range tree.Iter() {
		if err := a.analyse(p); err != nil {
			return err
		}
	}
	return nil
}

func (a *analyser) analyse(p model.Parsed) error {
	for _, name := range p.Packages() {
		pkg := p.Package(name)
		if strings.HasSuffix(name, "_test") {
			continue
		}
		for fileName, file := range pkg.Files {
			if strings.HasSuffix(fileName, "_test.go") {
				continue
			}
			a.analyseFile(fileName, file, p)
		}
	}
	return nil
}

type wRegistrar struct {
	packageName string
	fileName    string
	detected    []precedent
}

func (r *wRegistrar) register(pos token.Pos, i imported) {
	for _, l := range i.getLevel() {
		r.detected = append(r.detected, precedent{
			dependedOn:  i.getPackage(),
			packageName: r.packageName,
			fileName:    r.fileName,
			filePos:     pos,
			level:       l,
		})
	}
}

func (a *analyser) analyseFile(fileName string, f *ast.File, p model.Parsed) {
	var reg = wRegistrar{packageName: p.GetPath(), fileName: fileName}
	var v = visitor{
		walker: &reg,
		scope:  a.makeFileScope(f),
		node:   f,
	}
	ast.Walk(&v, f)
	a.precedents = append(a.precedents, reg.detected...)
}

func (a *analyser) makeFileScope(f *ast.File) []obj {
	var scope []obj
	for _, imp := range f.Imports {
		if imp.Path == nil {
			continue
		}
		packagePath := imp.Path.Value[1 : len(imp.Path.Value)-1]
		pkg := a.tree.Lookup(packagePath)
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
