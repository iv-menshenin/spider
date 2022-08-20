package importwalker

import (
	"context"
	"github.com/iv-menshenin/spider/importwalker/internal/model"
	"go/token"
	"log"
	"runtime/debug"

	"github.com/iv-menshenin/spider/importwalker/internal/fileparser"
	"github.com/iv-menshenin/spider/importwalker/internal/projparser"
)

type (
	Precedent struct {
		DependedOn  string
		PackageName string
		FileName    string
		FilePos     token.Pos
		Level       model.Level
	}
	Walker struct {
		mainPath    string
		projectPath string
		analyser    analyser
	}
)

func (w *Walker) Ident() string {
	return "walker"
}

func (w *Walker) Init(context.Context) error {
	defer func() {
		r := recover()
		if r != nil {
			log.Println(r)
			debug.PrintStack()
		}
	}()
	var parser = projparser.New(w.projectPath, w.mainPath, fileparser.New(fileparser.FilterExcludeTest))
	ast, err := parser.Parse()
	if err != nil {
		return err
	}
	if err = w.startAnalyse(ast); err != nil {
		return err
	}
	return nil
}

func (w *Walker) Ping(context.Context) error {
	return nil
}

func (w *Walker) Close() error {
	return nil
}

func New(projectPath, mainPath string) *Walker {
	return &Walker{
		mainPath:    mainPath,
		projectPath: projectPath,
	}
}
