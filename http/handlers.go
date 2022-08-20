package http

import "net/http"

func RegisterWWW(path string) {
	http.Handle("/web/", &Web{
		webPath: path,
	})
}
