package importwalker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

func (w *Walker) prepareInfo() error {
	modFilepath := path.Join(w.projectPath, "go.mod")
	data, err := loadModData(modFilepath)
	if err != nil {
		return err
	}
	f, err := modfile.Parse(modFilepath, data, nil)
	if err != nil {
		return err
	}
	w.enrichByGoMod(modFilepath, f)
	w.analyser.deps = w.depsExplore()
	for src, dep := range w.analyser.deps {
		for dst, pkg := range dep.deps {
			fmt.Printf("%s => %s [%d]\n", src, dst, len(pkg.prc))
		}
	}
	var cropper cropper
	for _, p := range w.analyser.deps["github.com/PetStores/go-simple/internal/resources"].deps {
		for _, d := range p.prc {
			fmt.Printf("***\ndependency: %s on [%s]; position: %d\n", d.Level, d.DependedOn, d.FilePos)
			s, err := cropper.cropFileExpr(d.FileName, int64(d.FilePos[0]), int64(d.FilePos[1]))
			fmt.Printf("%s\nerr: %+v\n\n", s, err)
		}
	}
	return nil
}

func loadModData(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (w *Walker) enrichByGoMod(modFilepath string, f *modfile.File) {
	w.analyser.modInfo.modPath = path.Dir(modFilepath)
	w.analyser.modInfo.module = f.Module.Mod.Path
	w.analyser.modInfo.goVersion = f.Go.Version
	for _, req := range f.Require {
		w.analyser.modInfo.require = append(
			w.analyser.modInfo.require,
			module{
				module:  req.Mod.Path,
				version: req.Mod.Version,
			},
		)
	}
	for _, rep := range f.Replace {
		w.analyser.modInfo.replace = append(
			w.analyser.modInfo.replace,
			[2]module{
				{module: rep.Old.Path, version: rep.Old.Version},
				{module: rep.New.Path, version: rep.New.Version},
			},
		)
	}
}

func (w *Walker) depsExplore() map[string]Dependency {
	var deps = make(map[string]Dependency)
	for _, prc := range w.analyser.precedents {
		var packageName = prc.PackagePath
		packageName = strings.TrimPrefix(packageName, w.projectPath)
		packageName = strings.TrimPrefix(packageName, "./")
		if !strings.HasPrefix(packageName, "vendor/") {
			packageName = path.Join(w.analyser.modInfo.module, packageName)
		}
		packageName = strings.TrimPrefix(packageName, "vendor/")
		dep, ok := deps[packageName]
		if !ok {
			dep.deps = make(map[string]Precedents)
		}
		prcs := dep.deps[prc.DependedOn]
		prcs.prc = append(prcs.prc, prc)
		dep.deps[prc.DependedOn] = prcs
		deps[packageName] = dep
	}
	return deps
}

type (
	Dependency struct {
		deps map[string]Precedents
	}
	Precedents struct {
		prc []Precedent
	}
)

func (c *cropper) cropFileExpr(fileName string, pos, posEnd int64) (string, error) {
	lines, err := c.openFile(fileName)
	if err != nil {
		return "", err
	}
	var start, end = 0, 0
	var dataL int
	for i, line := range lines {
		if line.pos > pos && start == 0 {
			if start = i - 3; start < 0 {
				start = 0
			}
			end = i + 2
		}
		if start > -1 {
			dataL += len(line.data)
			if dataL < 32 && i > end {
				end = i
			}
		}
	}

	if end == 0 || end > len(lines) {
		end = len(lines)
	}
	var result = []string{
		fmt.Sprintf("%s [%d:%d]\n", fileName, start+1, end+1),
	}
	var printed bool
	var lastPos int64
	for i, line := range lines[start:end] {
		if line.pos > pos && !printed {
			printed = true
			result = append(result, strings.Repeat(" ", 6+int(pos-lastPos))+strings.Repeat("~", int(posEnd-pos)+1))
		}
		lineNum := strconv.Itoa(start + i + 1)
		lineNum += strings.Repeat(" ", 6-len(lineNum))
		result = append(result, lineNum+string(line.data))
		lastPos = line.pos
	}
	return strings.Join(result, "\n"), nil
}

func (c *cropper) openFile(fileName string) ([]lineInfo, error) {
	if c.files == nil {
		c.files = make(map[string][]lineInfo)
	}
	if info, ok := c.files[fileName]; ok {
		return info, nil
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileData, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var (
		pos   int
		split = bytes.Split(fileData, []byte{'\n'})
		lines = make([]lineInfo, 0, len(split))
	)
	for i := range split {
		lines = append(lines, lineInfo{
			pos:  int64(pos),
			data: split[i],
		})
		pos += len(split[i]) + 1
	}
	c.files[fileName] = lines
	return lines, nil
}

type (
	cropper struct {
		files map[string][]lineInfo
	}
	lineInfo struct {
		pos  int64
		data []byte
	}
)
