package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
)

type RequestHeader struct {
	Method      string
	Path        string
	ContentType string
}

type Request struct {
	Header RequestHeader
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

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := parseRequest(conn)
}

func parseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)
	dat, err := reader.ReadString('\n')
	if err != nil {
		return &Request{}, err
	}

	part := strings.SplitN(strings.TrimSpace(dat), " ", 3)
	if len(part) != 3 {
		return &Request{}, fmt.Errorf("Invalid request format: %s. Usage: <Method> <Path> <Body>", dat)
	}

	header := RequestHeader{
		Method: part[0],
		Path:   part[1],
	}

	for {
		dat, err = reader.ReadString('\n')
		if err != nil {
			return &Request{}, err
		}
		dat = strings.TrimSpace(dat)
		if dat == "" {
			// End of headers / No content-type provided
			break
		}

		if strings.Contains(dat, "Content-Type") {
			ctnType := strings.SplitN(dat, ":", 2)
			if len(ctnType) == 2 {
				header.ContentType = strings.TrimSpace(ctnType[1])
			}
		}
	}

	bodyBytes := make([]byte, 0)
	tempBuffer := new(bytes.Buffer)
	_, err = io.Copy(tempBuffer, reader)
	if err != nil && err != io.EOF {
		return &Request{}, err
	}
	bodyBytes = tempBuffer.Bytes()

	return &Request{Header: header, Body: string(bodyBytes)}, nil
}
