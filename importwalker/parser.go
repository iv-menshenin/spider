package importwalker

import (
	"go/ast"
	"strings"
)

type (
	packageParser interface {
		Parse(name, path string) (map[string]*ast.Package, error)
	}
)

func (w *Walker) parse(name, path string) (packages, error) {
	p, err := w.parser.Parse(name, path)
	if err != nil {
		return nil, err
	}
	w.pkgs = append(w.pkgs, parsed{
		path: path,
		name: name,
		pkgs: p,
	})
	return p, nil
}

func (w *Walker) prepare() error {
	if packName := getCurrentPackageName(w.projectPath); packName != "" {
		w.pkgPaths[packName] = w.projectPath
	}
	for len(w.queue) > 0 {
		packName := ""
		currPack := w.queue[0]
		packageName := w.normalPackageName(currPack)
		pkgs, err := w.parse(currPack, packageName)
		if err != nil {
			return err
		}
		for pkgName, p := range pkgs {
			if packName == "" && !strings.HasSuffix(pkgName, "_test") {
				packName = pkgName
			}
			for _, file := range p.Files {
				for _, i := range file.Imports {
					packagePath := i.Path.Value
					if len(packagePath) < 3 {
						continue
					}
					packagePath = packagePath[1 : len(packagePath)-1]
					if isStdPackageName(packagePath) {
						continue
					}
					w.appendToQueue(packagePath)
				}
			}
		}
		if packName != "" {
			w.pkgNames[currPack] = packName
		}
		w.parsed = append(w.parsed, currPack)
		w.queue = w.queue[1:]
	}
	return nil
}

func (w *Walker) appendToQueue(packagePath string) {
	for _, q := range w.queue {
		if packagePath == q {
			return
		}
	}
	for _, p := range w.parsed {
		if p == packagePath {
			return
		}
	}
	w.queue = append(w.queue, packagePath)
}

func (w *Walker) normalPackageName(packagePath string) string {
	if strings.HasPrefix(packagePath, "./") {
		return packagePath
	}
	for pattern, path := range w.pkgPaths {
		if pattern == "" {
			continue
		}
		if strings.HasPrefix(packagePath, pattern) {
			return path + "/" + packagePath[len(pattern):]
		}
	}
	return "./vendor/" + packagePath
}

//func (w *Walker) findDeclaration(pkgName, pkgNames, objName string) []ast.Node {
//	for _,pkg := range w.pkgs {
//		if pkg.name == pkgName {
//			println(pkgNames)
//			for _, file := range pkg.pkgs[pkgNames].Files {
//				for _, decl := range file.Decls {
//					switch n := decl.(type) {
//					case *ast.FuncDecl:
//						if n.Name.Name == objName {
//							for _, rt := range n.Type.Results.List {
//								rt.Type
//							}
//						}
//					}
//				}
//			}
//		}
//	}
//	return nil
//}
