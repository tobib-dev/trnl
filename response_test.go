package trnl

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestWriteHeader(t *testing.T) {

	tests := []struct {
		name        string
		have        int
		expectPanic bool
	}{
		{"Status OK", StatusOK, false},
		{"Status Created", StatusCreated, false},
		{"Status Internal Server Error", StatusInternalServerError, false},
		{"Invalid code", 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			res := response{
				writer: bufio.NewWriter(&buf),
				header: Header{},
				status: tt.have,
			}

			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for status code: %d, but did not panic", tt.have)
						return
					}
				}()
			}

			res.WriteHeader(tt.have)
		})
	}
}

func TestWriteResponse(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		want string
	}{
		{"valid content", []byte("Hello, world!"), "HTTP/1.1 200 OK\r\n\r\nHello, world!"},
		{"empty content", []byte{}, "HTTP/1.1 200 OK\r\n\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			res := response{
				writer:      bufio.NewWriter(&buf),
				header:      Header{},
				protVer:     "HTTP/1.1",
				wroteHeader: false,
			}

			n, err := res.Write(tt.body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
				return
			}

			res.writer.Flush()

			msg := strings.Split(tt.want, "\r\n\r\n")

			if n != len(strings.TrimSpace(msg[1])) {
				t.Errorf("got - %d bytes to be written, want %d", n, len(tt.want))
				return
			}

			if buf.String() != tt.want {
				t.Errorf("got - %s, want %s", buf.String(), tt.want)
				return
			}
		})
	}
}
