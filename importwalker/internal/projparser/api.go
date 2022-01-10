package projparser

import (
	"go/ast"
	"sort"

	"github.com/iv-menshenin/spider/importwalker/internal/model"
)

func (w *AST) Len() int {
	return len(w.pkgs)
}

func (w *AST) Iter() <-chan model.Parsed {
	var ch = make(chan model.Parsed)
	go func() {
		defer close(ch)
		for i := range w.pkgs {
			ch <- w.pkgs[i]
		}
	}()
	return ch
}

func (w *AST) Lookup(packageName string) model.Parsed {
	for _, pkg := range w.pkgs {
		if pkg.name == packageName {
			return pkg
		}
	}
	return nil
}

func (p parsed) GetName() string {
	return p.name
}

func (p parsed) GetPath() string {
	return p.path
}

func (p parsed) Packages() []string {
	var names []string
	for name := range p.pkgs.ast {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (p parsed) Package(name string) *ast.Package {
	if name == "" {
		name = p.pkgs.alias
	}
	return p.pkgs.ast[name]
}
