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
	precedent struct {
		dependedOn  string
		packageName string
		fileName    string
		filePos     token.Pos
		level       model.Level
	}
	Walker struct {
		mainPath    string
		projectPath string
	}
)

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
