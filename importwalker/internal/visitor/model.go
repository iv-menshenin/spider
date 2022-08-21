package visitor

import (
	"go/ast"
	"go/token"

	"github.com/iv-menshenin/spider/importwalker/internal/model"
)

type (
	registrar interface {
		Register(token.Pos, token.Pos, model.Imported)
	}
	obj struct {
		name     string
		imported model.Imported
	}
	visitor struct {
		parent   *visitor
		node     ast.Node
		walker   registrar
		lookuper model.PackageLookuper

		scope []obj
	}
)
