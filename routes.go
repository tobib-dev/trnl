package main

import (
	"sync/atomic"
)

var muxIDCounter uint64

type routeKey struct {
	method string
	path   string
}

type Mux struct {
	id     uint64
	routes map[routeKey]Handler
}

func (m *Mux) ServeHTTP(w ResponseWriter, r *Request) {
	hdlr, ok := m.routes[routeKey{r.Header.Method, r.Header.Path}]
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte("404 Not Found"))
		return
	}
	hdlr.ServeHTTP(w, r)
}

func Default() *Mux {
	return &Mux{
		id:     atomic.AddUint64(&muxIDCounter, 1),
		routes: make(map[routeKey]Handler),
	}
}

func (m *Mux) Get(path string, handler Handler) {
	m.routes[routeKey{method: "GET", path: path}] = handler
}

func (m *Mux) Post(path string, handler Handler) {
	m.routes[routeKey{method: "POST", path: path}] = handler
}

func (m *Mux) Put(path string, handler Handler) {
	m.routes[routeKey{method: "PUT", path: path}] = handler
}

func (m *Mux) Delete(path string, handler Handler) {
	m.routes[routeKey{method: "DELETE", path: path}] = handler
}
