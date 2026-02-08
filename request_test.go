package trnl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

type mockConn struct {
	net.Conn
	buffer *bytes.Buffer
}

func (m *mockConn) Read(b []byte) (n int, e error) {
	return m.buffer.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, e error) {
	return m.buffer.Write(b)
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockConn) Close() error {
	return nil
}

func readersEqual(r1, r2 io.Reader) (bool, error) {
	if r1 == nil && r2 == nil {
		return true, nil
	}

	if r1 == nil || r2 == nil {
		return false, nil
	}

	b1, err := io.ReadAll(r1)
	if err != nil {
		return false, err
	}

	b2, err := io.ReadAll(r2)
	if err != nil {
		return false, err
	}

	return bytes.Equal(b1, b2), nil
}

func TestParseRequest(t *testing.T) {

	// Valid request
	content := "POST /users HTTP/1.1\r\nContent-Type: application/json\r\nContent-Length: 18\r\n\r\n{'data': 'ABC123'}"

	vConn := &mockConn{
		buffer: new(bytes.Buffer),
	}
	fmt.Fprint(vConn, content)

	body := []byte("{'data': 'ABC123'}")
	buf := bytes.NewBuffer(body)

	vReq := Request{
		Header: RequestHeader{
			Method:          "POST",
			Path:            "/users",
			ContentType:     "application/json",
			ContentLength:   18,
			ProtocolVersion: "HTTP/1.1",
		},
		Body: buf,
	}

	// Invalid request - no protocol version
	noProtCnt := "GET /users\r\nContent-Type: application/json\r\nContent-Length: 0\r\n\r\n"
	noProtVer := &mockConn{
		buffer: new(bytes.Buffer),
	}
	fmt.Fprint(noProtVer, noProtCnt)
	npvMsg := fmt.Sprintf("Invalid request format: %s. Usage: <Method> <Path> <metadata>",
		"GET /users\r\n")

	// Invalid request - no http method
	noMethodCnt := "/users HTTP/1.1\r\nContent-Type: application/json\r\nContent-Length: 0\r\n\r\n"
	noMethod := &mockConn{
		buffer: new(bytes.Buffer),
	}
	fmt.Fprint(noMethod, noMethodCnt)
	nmMsg := fmt.Sprintf("Invalid request format: %s. Usage: <Method> <Path> <metadata>",
		"/users HTTP/1.1\r\n")

	var tests = []struct {
		name string
		have net.Conn
		want Request
		e    error
	}{
		{"valid request", vConn, vReq, nil},
		{"no http version", noProtVer, Request{
			Header: RequestHeader{},
			Body:   bytes.NewBuffer(nil),
		}, errors.New(npvMsg)},
		{"no http method", noMethod, Request{
			Header: RequestHeader{},
			Body:   bytes.NewBuffer(nil),
		}, errors.New(nmMsg)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.have.Close()

			tt.have.SetReadDeadline(time.Now().Add(time.Second))
			res, err := parseRequest(tt.have)
			if (err == nil) && (tt.e != nil) {
				t.Errorf("test failed - got error %v, want error %v", err, tt.e)
				return
			} else if err != nil && tt.e != nil && err.Error() != tt.e.Error() {
				t.Errorf("test failed - got %v, want %v", err, tt.e)
				return
			}

			if tt.e == nil && res == nil {
				t.Fatalf("parseRequest returned nil res when no error was expected")
			}
			if res != nil {
				if res.Header.Method != tt.want.Header.Method ||
					res.Header.Path != tt.want.Header.Path ||
					res.Header.ContentLength != tt.want.Header.ContentLength ||
					res.Header.ContentType != tt.want.Header.ContentType ||
					res.Header.ProtocolVersion != tt.want.Header.ProtocolVersion {
					t.Errorf("test failed - got Header: %v, want Header: %v", &res.Header, tt.want.Header)
					return
				}

				eq, err := readersEqual(res.Body, tt.want.Body)
				if err != nil && !eq {
					t.Errorf("test failed - got %v, want %v", &res.Body, tt.want)
					return
				}
			}
		})
	}
}
