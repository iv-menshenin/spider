package importwalker

import (
	"context"
	"github.com/iv-menshenin/spider/importwalker/parser"
	"go/ast"
	"go/token"
	"log"
	"runtime/debug"
)

type (
	level    int
	packages map[string]*ast.Package
	parsed   struct {
		path string
		name string
		pkgs packages
	}
	precedent struct {
		dependedOn  string
		packageName string
		fileName    string
		filePos     token.Pos
		level       level
	}
	Walker struct {
		pkgs   []parsed
		queue  []string
		parsed []string

		pkgNames    map[string]string
		pkgPaths    map[string]string
		projectPath string

		parser     packageParser
		precedents []precedent
	}
)

const (
	levelImportNone level = iota
	levelImportMethod
	levelImportFunc
	levelImportStruct
)

func (w *Walker) Init(context.Context) error {
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r)
			debug.PrintStack()
		}
	}()
	if err := w.prepare(); err != nil {
		return err
	}
	if err := w.startAnalyse(); err != nil {
		return err
	}
	return nil
}

func (w *Walker) Ping(ctx context.Context) error {
	return nil
}

func (w *Walker) Close() error {
	return nil
}

func New(projectPath, mainPath string) *Walker {
	var pkgPaths = make(map[string]string)
	pkgPaths[""] = mainPath
	return &Walker{
		queue:       []string{mainPath},
		pkgPaths:    pkgPaths,
		pkgNames:    make(map[string]string),
		projectPath: projectPath,
		parser:      parser.New(parser.FilterExcludeTest),
	}
}
