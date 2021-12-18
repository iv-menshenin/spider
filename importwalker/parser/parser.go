package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type (
	Parser struct {
		fs *token.FileSet
	}
)

func (w *Parser) Parse(name, path string) (map[string]*ast.Package, error) {
	p, err := parser.ParseDir(w.fs, path, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func New() *Parser {
	return &Parser{
		fs: token.NewFileSet(),
	}
}
