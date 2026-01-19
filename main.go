package main

import (
	"fmt"
	"net"
	"strings"
)

type Request struct {
	Header string
	Body   string
}

type Response struct {
	Header  string
	Status  int
	Payload interface{}
}

type ResponseWriter interface {
	Header() string
	WriteHeader(status int)
	Write(p []byte) (int, error)
}

type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }

type routeKey struct {
	method string
	path   string
}

type Mux struct {
	routes map[routeKey]Handler
}

type Listener struct {
	net.Listener
}

type Server struct {
	Addr    string
	Handler Handler
	//ReadTimeout  time.Duration
	//WriteTimeout time.Duration
	//IdleTimeout  time.Duration
}

func Default() Mux {
	return Mux{
		routes: make(map[routeKey]Handler),
	}
}

func (s *Server) ListenAndServe() error {
	addr := strings.Split(s.Addr, ":")

	if len(addr) != 2 {
		return fmt.Errorf("Invalid address format! Usage <host>:<port>")
	}

	port := fmt.Sprintf(":%s", addr[1])

	if s.Handler == nil {
		return fmt.Errorf("Handler is nil")
	}

	list, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer list.Close()

	return s.Serve(list)
}

func (s *Server) Serve(l net.Listener) error
