package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }

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

func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := parseRequest(conn)
	if err != nil {
		return
	}

	res := &response{
		conn:   conn,
		writer: bufio.NewWriter(conn),
	}

	s.Handler.ServeHTTP(res, req)
}
