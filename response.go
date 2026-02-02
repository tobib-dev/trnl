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

/*
 * Write Header to response
 *
 * For the case of this being a personal project, I have not included all
 * HTTP response codes. Hence, WriteHeader may panic if the given code
 * is not in already included in the status_codes.go file.
 */
func (r *response) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true

	msg, ok := StatusMessages[status]
	if !ok {
		log.Printf("Invalid status code: %d\n", status)
		panic(fmt.Sprintf("Invalid HTTP status code: %d. Please use a valid HTTP resposne code.", status))
	}
	// Write response status line
	fmt.Fprintf(r.writer, "HTTP/1.1 %d %s\r\n", status, msg)

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
