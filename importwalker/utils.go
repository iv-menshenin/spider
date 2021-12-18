package importwalker

import (
	"bufio"
	"io"
	"os"
	"runtime"
	"strings"
)

func getCurrentPackageName(path string) string {
	f, err := os.Open(path + "/go.mod")
	if err != nil {
		return ""
	}
	defer f.Close()
	return grabModuleNameFromGoMod(f)
}

func grabModuleNameFromGoMod(r io.Reader) string {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parsed := strings.Fields(scanner.Text())
		if len(parsed) == 2 && parsed[0] == "module" {
			return parsed[1]
		}
	}
	return ""
}

func isStdPackageName(path string) bool {
	return directoryExists(runtime.GOROOT()+"/src/"+path) || path == "C"
}

func directoryExists(path string) bool {
	info, e := os.Stat(path)
	if e == nil {
		return info.IsDir()
	}
	return false
}
