package importwalker

import (
	"github.com/iv-menshenin/spider/importwalker/internal/model"
	"github.com/iv-menshenin/spider/importwalker/internal/visitor"
	"go/ast"
	"go/token"
	"strings"
)

type (
	analyser struct {
		tree          model.AST
		precedents    []precedent
		filesAnalysed []string
	}
)

func (w *Walker) startAnalyse(tree model.AST) error {
	if tree.Len() == 0 {
		return nil
	}
	w.analyser = analyser{
		tree:       tree,
		precedents: []precedent{},
	}
	for p := range tree.Iter() {
		if err := w.analyser.analyse(p); err != nil {
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

func (r *wRegistrar) Register(pos token.Pos, i model.Imported) {
	for _, l := range i.GetLevel() {
		r.detected = append(r.detected, precedent{
			dependedOn:  i.GetPackage(),
			packageName: r.packageName,
			fileName:    r.fileName,
			filePos:     pos,
			level:       l,
		})
	}
}

func (a *analyser) analyseFile(fileName string, f *ast.File, p model.Parsed) {
	var reg = wRegistrar{packageName: p.GetPath(), fileName: fileName}
	var v = visitor.New(f, &reg, a.tree)
	ast.Walk(v, f)
	a.precedents = append(a.precedents, reg.detected...)
	a.filesAnalysed = append(a.filesAnalysed, fileName)
}
