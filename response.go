package trnl

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type Header map[string][]string

func (h Header) Set(key, value string) {
	h[key] = []string{value}
}

func (h Header) Add(key, value string) {
	h[key] = append(h[key], value)
}

func (h Header) Write(w io.Writer) error {
	// Write header key and values to response header
	for k, v := range h {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

type response struct {
	conn        net.Conn
	writer      *bufio.Writer
	header      Header
	status      int
	protVer     string
	wroteHeader bool
}

type ResponseWriter interface {
	Header() Header
	WriteHeader(status int)
	Write(p []byte) (int, error)
}

// Implement the flusher interface on response type
type Flusher interface {
	Flush() error
}

// Set response Header
func (r *response) Header() Header {
	return r.header
}

// Write Header to response
func (r *response) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true

	// Write protocol version and status
	fmt.Fprintf(r.writer, "%s: %d\r\n", r.protVer, status)

	// Write headers
	if err := r.header.Write(r.writer); err != nil {
		log.Printf("error writing header: %v", err)
		return
	}

	r.writer.WriteString("\r\n")
}

// Write data to response body
func (r *response) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(StatusOK)
	}
	return r.writer.Write(p)
}

func (r *response) Flush() error {
	if !r.wroteHeader {
		r.WriteHeader(200)
	}
	return r.writer.Flush()
}
