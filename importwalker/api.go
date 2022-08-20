package importwalker

import (
	"io"
	"os"
)

func (w *Walker) FileList() []string {
	return w.analyser.filesAnalysed
}

func (w *Walker) FileContent(id int) (string, error) {
	if id < len(w.analyser.filesAnalysed) && id >= 0 {
		f, err := os.Open(w.analyser.filesAnalysed[id])
		if err != nil {
			return "", err
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", io.EOF
}

func (w *Walker) FileDeps(id int) ([]Precedent, error) {
	if id >= len(w.analyser.filesAnalysed) || id < 0 {
		return nil, io.EOF
	}
	var result []Precedent
	name := w.analyser.filesAnalysed[id]
	for _, p := range w.analyser.precedents {
		if p.FileName == name {
			result = append(result, p)
		}
	}
	return result, nil
}
