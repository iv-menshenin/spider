package importwalker

import (
	"go/ast"
	"strings"
)

type (
	packageParser interface {
		Parse(path string) (map[string]*ast.Package, error)
	}
)

func (w *Walker) parseProject() error {
	if packName := getCurrentPackageName(w.projectPath); packName != "" {
		w.pkgPaths[packName] = w.projectPath
	}
	for len(w.queue) > 0 {
		packageName := w.queue[0]
		packagePath := w.detectPackagePath(packageName)
		if err := w.parsePackage(packageName, packagePath); err != nil {
			return err
		}
		w.parsed = append(w.parsed, packageName)
		w.queue = w.queue[1:]
	}
	return nil
}

func (w *Walker) parsePackage(name, path string) error {
	parsedPackages, err := w.parser.Parse(path)
	if err != nil {
		return err
	}
	w.pkgs = append(w.pkgs, parsed{
		path: path,
		name: name,
		pkgs: parsedPackages,
	})
	var packName string
	for pkgName, p := range parsedPackages {
		if packName == "" && !strings.HasSuffix(pkgName, "_test") {
			packName = pkgName
		}
		for _, file := range p.Files {
			w.addImportsToQueue(file.Imports)
		}
	}
	if packName != "" {
		w.pkgNames[name] = packName
	}
	return nil
}

func (w *Walker) addImportsToQueue(imports []*ast.ImportSpec) {
	for _, i := range imports {
		importedName := i.Path.Value
		if len(importedName) < 3 {
			continue
		}
		importedName = importedName[1 : len(importedName)-1]
		if isStdPackageName(importedName) {
			continue
		}
		w.appendToQueue(importedName)
	}
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

// detectPackagePath defines the relationship between the full package name and its location in the file system.
// Returns the full path to the specified package files in the file system if the package was found in the w.pkgPaths registry.
// If no package was found, it tries to suggest a relative path through the `vendor` directory
func (w *Walker) detectPackagePath(fullPackagePath string) string {
	if strings.HasPrefix(fullPackagePath, "./") {
		return fullPackagePath
	}
	for knownPackageName, knownPackagePath := range w.pkgPaths {
		if knownPackageName == "" {
			continue
		}
		if strings.HasPrefix(fullPackagePath, knownPackageName) {
			return knownPackagePath + "/" + fullPackagePath[len(knownPackageName):]
		}
	}
	return "./vendor/" + fullPackagePath
}
