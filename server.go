package trnl

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
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
	Addr        string
	Handler     Handler
	ReadTimeout time.Duration
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

		// Set Server ReadTimeout only if positive
		if s.ReadTimeout > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(s.ReadTimeout)); err != nil {
				log.Printf("failed to set read deadline (timeout=%v, remote=%s): %v", s.ReadTimeout, conn.RemoteAddr(), err)
				conn.Close()
				continue
			}
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
		conn:    conn,
		writer:  bufio.NewWriter(conn),
		header:  make(Header),
		protVer: "HTTP/1.1",
	}

	req, err := parseRequest(conn)
	if err != nil {
		// Respond with HTTP Status BadRequest and flush response
		res.WriteHeader(StatusBadRequest)
		res.Flush()
		return
	}

	res.header.Set("location", req.Header.Path)
	s.Handler.ServeHTTP(res, req)
	res.Flush()
}
