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

/*
 * ListenAndServe starts a server listening on the given address and handler.
 */
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

/*
 * Serves a Server type using a net.Listener interface as parameter
 */
func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}
}

/*
 * Handle connections on listeners. Respond with bad request if error
 * is encoutered error while parsing. Else, serve request if valid
 */
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	res := &response{
		conn:   conn,
		writer: bufio.NewWriter(conn),
	}

	req, err := parseRequest(conn)
	if err != nil {
		// Respond with HTTP Status BadRequest
		res.WriteHeader(400)
		return
	}

	s.Handler.ServeHTTP(res, req)
}
