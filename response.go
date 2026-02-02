package trnl

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
)

type Header map[string][]string

func (h Header) Set(key, value string) {
	h[key] = append(h[key], value)
}

func (h Header) Write(w io.Writer, status int, protVer string) error {
	// Write Protocol version and status to response header
	verAndStatus := fmt.Sprintf(protVer, strconv.Itoa(status))
	_, err := w.Write([]byte(verAndStatus))
	if err != nil {
		return err
	}

	// Write header key and values to response header
	for k, v := range h {
		keyVal := fmt.Sprintf("%s: %s", k, v)
		_, err := w.Write([]byte(keyVal))
		if err != nil {
			return err
		}
	}

	return nil
}

type response struct {
	conn     net.Conn
	writer   *bufio.Writer
	header   Header
	status   int
	protVer  string
	endpoint string
}

type ResponseWriter interface {
	Header() Header
	WriteHeader(status int)
	Write(p []byte) (int, error)
}

// Set response Header
func (r *response) Header() Header {
	return r.header
}

// Write Header to response
func (r *response) WriteHeader(status int) {
	err := r.header.Write(r.writer, status, r.protVer)
	if err != nil {
		r.resetResponsePayload()
	}
}

// Reset conn writer if write errors
func (r *response) resetResponsePayload() {
	r.writer.Reset(r.writer)
}

// Write data to response body
func (r *response) Write(p []byte) (int, error) {
	return r.writer.Write(p)
}
