package importwalker

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/iv-menshenin/spider/importwalker/internal/model"
	"github.com/iv-menshenin/spider/importwalker/internal/visitor"
)

type (
	analyser struct {
		modInfo       modInfo
		tree          model.AST
		precedents    []Precedent
		filesAnalysed []string
		deps          map[string]Dependency
	}
	modInfo struct {
		modPath   string
		module    string
		goVersion string
		require   []module
		replace   [][2]module
	}
	module struct {
		module  string
		version string
	}
)

func (w *Walker) startAnalyse(tree model.AST) error {
	if tree.Len() == 0 {
		return nil
	}
	w.analyser = analyser{
		tree:       tree,
		precedents: []Precedent{},
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
	packagePath string
	fileName    string
	detected    []Precedent
}

func (r *wRegistrar) Register(pos, end token.Pos, i model.Imported) {
	for _, l := range i.GetLevel() {
		r.detected = append(r.detected, Precedent{
			DependedOn:  i.GetPackage(),
			PackageName: r.packageName,
			PackagePath: r.packagePath,
			FileName:    r.fileName,
			FilePos:     [2]token.Pos{pos, end},
			Level:       l,
		})
	}
}

func (a *analyser) analyseFile(fileName string, f *ast.File, p model.Parsed) {
	var reg = wRegistrar{
		packageName: f.Name.Name,
		packagePath: p.GetPath(),
		fileName:    fileName,
	}
	var v = visitor.New(f, &reg, a.tree)
	ast.Walk(v, f)
	a.precedents = append(a.precedents, reg.detected...)
	a.filesAnalysed = append(a.filesAnalysed, fileName)
}
