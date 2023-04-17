package web

import "net/http"

type Namespace struct {
	path string
	server Server
}

func (h *Namespace) Get(path string, handleFunc HandleFunc) {
	h.server.addRoute(http.MethodGet, h.path + path, handleFunc)
}

func (h *Namespace) Namespace(path string) *Namespace {
	return &Namespace{
		server: h.server,
		path: h.path + path,
	}
}

// func NewNamespace(h *HTTPServer, path string) *Namespace {
// 	return &Namespace{
// 		server: h,
// 		path: path,
// 	}
// }
