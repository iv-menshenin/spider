package visitor

import (
	"go/ast"

	"github.com/iv-menshenin/spider/importwalker/internal/model"
)

type (
	importedFunc struct {
		fil importedFile
		fun *ast.FuncDecl
	}
)

func (p importedFunc) GetField(string) model.Imported {
	return nil
}

func (p importedFunc) GetResult(nn int) model.Imported {
	if p.fun.Type.Results == nil {
		return nil
	}
	var i = 0
	for _, ret := range p.fun.Type.Results.List {
		if len(ret.Names) == 0 {
			if i == nn {
				return p.fil.reference(ret.Type)
			}
			i++
		}
		for range ret.Names {
			if i == nn {
				return p.fil.reference(ret.Type)
			}
			i++
		}
	}
	return nil
}

func (p importedFunc) GetLevel() []model.Level {
	var result []model.Level
	if p.fun.Recv != nil {
		result = append(result, model.LevelImportMethod)
	} else {
		result = append(result, model.LevelImportFunc)
	}
	// todo: check returned types and add LevelImportStruct
	return result
}

func (p importedFunc) GetPackage() string {
	return p.fil.pkg.path
}
