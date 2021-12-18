package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iv-menshenin/appctl"
	"github.com/iv-menshenin/spider/importwalker"
)

type (
	application struct {
		args     Args
		basePath string
		walker   *importwalker.Walker
	}
)

func main() {
	spider := makeApplication()
	services := appctl.ServiceKeeper{
		Services: []appctl.Service{
			spider.walker,
		},
	}
	app := appctl.Application{
		MainFunc:              spider.mainFunc,
		Resources:             &services,
		TerminationTimeout:    time.Second * 5,
		InitializationTimeout: time.Second * 30,
	}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func makeApplication() application {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	args, err := parseArgs()
	if err != nil {
		panic(err)
	}
	return application{
		args:     args,
		basePath: wd,
		walker:   importwalker.New(wd, args.ParseFile),
	}
}

const startPage = "http://localhost:28080/web/index.html"

func (a *application) mainFunc(_ context.Context, halt <-chan struct{}) error {
	a.registerHandlers()
	go func() {
		<-time.After(time.Second)
		if err := browser(startPage); err != nil {
			log.Println(err)
		}
	}()
	if a.args.ParseOnly {
		return nil
	}
	return startHttp(halt)
}

func (a *application) registerHandlers() {
	//http.Handle("/import/imports-net", &Net{
	//	graphGetter: a.walker,
	//})
	//http.Handle("/import/imports-arcs", &ArcLinks{
	//	mainNode:    a.packageName,
	//	graphGetter: a.walker,
	//})
	//http.Handle("/code/", &File{
	//	basePath: a.basePath,
	//	analyser: a.walker,
	//})
	http.Handle("/web/", &Web{})
}
