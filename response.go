package main

import (
	"bufio"
	"net"
)

type response struct {
	conn   net.Conn
	writer *bufio.Writer
	header string
	status int
}

type ResponseWriter interface {
	Header() string
	WriteHeader(status int)
	Write(p []byte) (int, error)
}

func (r *response) Header() string {
	return r.header
}

func (r *response) WriteHeader(status int) {
	r.status = status
}

func (r *response) Write(p []byte) (int, error) {
	return r.writer.Write(p)
}
