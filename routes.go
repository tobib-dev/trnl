package trnl

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

/*
 * Serve request using method and route, respond with 404 if endpoint is not found
 */
func (m Mux) ServeHTTP(w ResponseWriter, r *Request) {
	hdlr, ok := m.routes[routeKey{r.Header.Method, r.Header.Path}]
	if !ok {
		w.WriteHeader(StatusNotFound)
		w.Write([]byte("404 Not Found"))
		return
	}

	hdlr.ServeHTTP(w, r)
}

/*
 * Generate and return default multiplexer or mini-router
 */
func Default() Mux {
	return Mux{
		id:     atomic.AddUint64(&muxIDCounter, 1),
		routes: make(map[routeKey]Handler),
	}
}

/*
 * Add Get requests multiplexer
 */
func (m *Mux) Get(path string, handler any) {
	m.handle("GET", path, handler)
}

/*
 * Add Post requests to multiplexer
 */
func (m *Mux) Post(path string, handler any) {
	m.handle("POST", path, handler)
}

/*
 * Add Put/Update requests to multiplexer
 */
func (m *Mux) Put(path string, handler any) {
	m.handle("PUT", path, handler)
}

/*
 * Add delete request to multiplexer
 */
func (m *Mux) Delete(path string, handler any) {
	m.handle("DELETE", path, handler)
}

/*
 * Map methods and paths to handler functions
 */
func (m Mux) handle(method, path string, handler any) {
	switch h := handler.(type) {
	case Handler:
		m.routes[routeKey{method: method, path: path}] = h
	case func(ResponseWriter, *Request):
		m.routes[routeKey{method: method, path: path}] = HandlerFunc(h)
	default:
		panic("Invalid handler type, handler must be Handler interface or func(ResponseWriter, *Request)")
	}
}
