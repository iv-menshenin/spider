package fileparser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"strings"
)

type (
	FilterFunc func(fs.FileInfo) bool
	Parser     struct {
		filterFn FilterFunc
	}
)

// FilterExcludeTest filters out test files from packages
func FilterExcludeTest(fi fs.FileInfo) bool {
	return !strings.HasSuffix(fi.Name(), "_test.go")
}

// FilterIncludeAll allows you to get all files without any filtering
func FilterIncludeAll(fs.FileInfo) bool {
	return true
}

func (w *Parser) Parse(packagePath string) (map[string]*ast.Package, error) {
	list, err := os.ReadDir(packagePath)
	if err != nil {
		return nil, err
	}

	var pkg = newPackage()
	for _, file := range list {
		if file.IsDir() {
			continue
		}
		if path.Ext(file.Name()) != ".go" {
			continue
		}

		var (
			src      *ast.File
			fileName = path.Join(packagePath, file.Name())
		)
		src, err = parser.ParseFile(token.NewFileSet(), fileName, nil, parser.ParseComments|parser.AllErrors)
		if err != nil {
			return nil, err
		}
		pkg.append(src.Name.Name, fileName, src)
	}

	return pkg.get(), nil
}

type Package struct {
	pkg map[string]*ast.Package
}

func (p *Package) append(pkgName string, fileName string, src *ast.File) {
	pkg, found := p.pkg[pkgName]
	if !found {
		pkg = &ast.Package{
			Name:  pkgName,
			Files: make(map[string]*ast.File),
		}
		p.pkg[pkgName] = pkg
	}
	pkg.Files[fileName] = src
}

func (p *Package) get() map[string]*ast.Package {
	return p.pkg
}

func newPackage() *Package {
	return &Package{
		pkg: make(map[string]*ast.Package),
	}
}

func New(filter FilterFunc) *Parser {
	return &Parser{
		filterFn: filter,
	}
}
