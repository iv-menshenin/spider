package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strings"
)

type (
	FilterFunc func(fs.FileInfo) bool
	Parser     struct {
		filterFn FilterFunc
		fs       *token.FileSet
	}
)

func FilterExcludeTest(fi fs.FileInfo) bool {
	return !strings.HasSuffix(fi.Name(), "_test.go")
}

func FilterIncludeAll(fi fs.FileInfo) bool {
	return true
}

func (w *Parser) Parse(path string) (map[string]*ast.Package, error) {
	p, err := parser.ParseDir(w.fs, path, w.filterFn, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func New(filter FilterFunc) *Parser {
	return &Parser{
		filterFn: filter,
		fs:       token.NewFileSet(),
	}
}
