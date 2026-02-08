package trnl

import (
	"bytes"
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
	content := "POST /users HTTP/1.1\nContent-Type: application/json\nContent-Length: 18\n\n{'data': 'ABC123'}"

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

	// Invalid request
	noProtContent := "GET /users"
	noProtVer := &mockConn{
		buffer: new(bytes.Buffer),
	}
	fmt.Fprint(noProtVer, noProtContent)
	//npvMsg := fmt.Sprintf("Invalid request format: %s. Usage: <Method> <Path> <metadata>", invalidContent)

	var tests = []struct {
		name string
		have net.Conn
		want Request
		e    error
	}{
		{"valid request", vConn, vReq, nil},
		//{"no http version", noProtVer, Request{}, errors.New(npvMsg)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.have.Close()

			tt.have.SetReadDeadline(time.Now().Add(time.Second))
			res, err := parseRequest(tt.have)
			if err != nil && err.Error() != tt.e.Error() {
				t.Errorf("test failed - got %v, want %v", err, tt.e)
			}
			if res != nil {
				if res.Header.Method != tt.want.Header.Method ||
					res.Header.Path != tt.want.Header.Path ||
					res.Header.ContentLength != tt.want.Header.ContentLength ||
					res.Header.ContentType != tt.want.Header.ContentType ||
					res.Header.ProtocolVersion != tt.want.Header.ProtocolVersion {
					t.Errorf("test failed - got Header: %v, want Header: %v", &res.Header, tt.want.Header)
				}

				eq, err := readersEqual(res.Body, tt.want.Body)
				if err != nil || !eq {
					t.Errorf("test failed - got %v, want %v", &res.Body, tt.want)
				}
			}
		})
	}
}
