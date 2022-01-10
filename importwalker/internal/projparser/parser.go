package projparser

import (
	"go/ast"
	"strings"
)

type (
	AST struct {
		pkgPaths map[string]string
		pkgs     []parsed
	}
	ProjectParser struct {
		projectPath string
		mainPath    string
		queue       []string
		parsed      []string
		parser      Parser
	}
	parsed struct {
		path string
		name string
		pkgs packages
	}
	Parser interface {
		Parse(path string) (map[string]*ast.Package, error)
	}
	packages struct {
		alias string
		ast   map[string]*ast.Package
	}
)

func (w *ProjectParser) Parse() (*AST, error) {
	var result = AST{
		pkgPaths: map[string]string{"": w.mainPath},
		pkgs:     nil,
	}
	w.clean()
	if packName := getCurrentPackageName(w.projectPath); packName != "" {
		result.pkgPaths[packName] = w.projectPath
	}
	for len(w.queue) > 0 {
		packageName := w.queue[0]
		packagePath := result.detectPackagePath(packageName)
		parsedPackages, err := w.parsePackage(packagePath)
		if err != nil {
			return nil, err
		}
		result.pkgs = append(result.pkgs, parsed{
			path: packagePath,
			name: packageName,
			pkgs: parsedPackages,
		})
		w.parsed = append(w.parsed, packageName)
		w.queue = w.queue[1:]
	}
	return &result, nil
}

func (w *ProjectParser) clean() {
	w.queue = []string{w.mainPath}
	w.parsed = []string{}
}

func (w *ProjectParser) parsePackage(path string) (packages, error) {
	parsedPackages, err := w.parser.Parse(path)
	if err != nil {
		return packages{}, err
	}
	var alias string
	for pkgName, p := range parsedPackages {
		if alias == "" && !strings.HasSuffix(pkgName, "_test") {
			alias = pkgName
		}
		for _, file := range p.Files {
			w.addImportsToQueue(file.Imports)
		}
	}
	return packages{alias: alias, ast: parsedPackages}, nil
}

func (w *ProjectParser) addImportsToQueue(imports []*ast.ImportSpec) {
	for _, i := range imports {
		importedName := i.Path.Value
		if len(importedName) < 3 {
			continue
		}
		importedName = importedName[1 : len(importedName)-1]
		if isStdPackageName(importedName) {
			continue
		}
		w.appendToQueueIfNotExists(importedName)
	}
}

func (w *ProjectParser) appendToQueueIfNotExists(packagePath string) {
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
func (w *AST) detectPackagePath(fullPackagePath string) string {
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

func New(projectPath, mainPath string, parser Parser) *ProjectParser {
	return &ProjectParser{
		mainPath:    mainPath,
		projectPath: projectPath,
		parser:      parser,
	}
}
