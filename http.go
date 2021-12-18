package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type (
	importGraphGetter interface {
		GetLinks(scope []string, format, sep string, writer io.Writer) error
		GetNodes(scope []string, format, sep string, writer io.Writer) error
	}
	Web struct{}
	Net struct {
		graphGetter importGraphGetter
	}
	ArcLinks struct {
		mainNode    string
		graphGetter importGraphGetter
	}
	packageFileAnalyser interface {
		GetDefaultPackageName(path string) string
	}
	File struct {
		basePath string
		analyser packageFileAnalyser
	}
)

func startHttp(stop <-chan struct{}) error {
	server := &http.Server{Addr: ":28080"}
	chErr := make(chan error, 2)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			chErr <- err
		}
	}()
	go func() {
		<-stop
		if err := server.Shutdown(context.Background()); err != nil {
			chErr <- err
		}
	}()
	if err, ok := <-chErr; ok {
		return err
	}
	return nil
}

func (f *Web) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("." + r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	switch path.Ext(r.URL.Path) {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	}
	if _, err = io.Copy(w, file); err != nil {
		fmt.Println(err)
	}
}

func (f *File) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimPrefix(r.URL.Path, "/code/")
	code, err := parser.ParseFile(token.NewFileSet(), f.basePath+filePath, nil, parser.ParseComments|parser.AllErrors)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ast.Walk(nil, code.Decls[0])
	file, err := os.Open(f.basePath + filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", "text/plain")
	if _, err = io.Copy(w, file); err != nil {
		fmt.Println(err)
	}
}

func (n *Net) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/csv")
	if _, err := w.Write([]byte("source,target,type\n")); err != nil {
		log.Println(err)
		return
	}
	if err := n.graphGetter.GetLinks(r.URL.Query()["option"], "{{ .Source }},{{ .Target }},{{ .Type }}", "", w); err != nil {
		log.Println(err)
	}
}

func (a *ArcLinks) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte("{\n  \"nodes\": [\n")); err != nil {
		log.Println(err)
		return
	}
	if _, err := w.Write([]byte("    {\"id\": \"" + a.mainNode + "\", \"group\": -1},\n")); err != nil {
		log.Println(err)
		return
	}
	if err := a.graphGetter.GetNodes(r.URL.Query()["option"], "    {\"id\": \"{{ .Name }}\", \"group\": {{ .Group }}}", ",", w); err != nil {
		log.Println(err)
	}
	if _, err := w.Write([]byte("  ],\n  \"links\": [\n")); err != nil {
		log.Println(err)
		return
	}
	if err := a.graphGetter.GetLinks(r.URL.Query()["option"], "    {\"source\": \"{{ .Source }}\", \"target\": \"{{ .Target }}\", \"value\": {{ .Count }}}", ",", w); err != nil {
		log.Println(err)
	}
	if _, err := w.Write([]byte("  ]\n}\n")); err != nil {
		log.Println(err)
		return
	}
}
