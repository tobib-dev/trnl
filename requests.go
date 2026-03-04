package trnl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type RequestHeader struct {
	Method          string
	Path            string
	ContentType     string
	ContentLength   int
	ProtocolVersion string
}

type Request struct {
	Header RequestHeader
	Body   io.Reader
	params map[string]string
}

/*
 * Return request parameter value and nil if available
 * else return empty string and error
 */
func (r *Request) Params(key string) (string, error) {
	prm, ok := r.params[key]
	if !ok {
		return "", fmt.Errorf("no parameters with the key: %s", key)
	}
	return prm, nil
}

/*
 * parseRequest handles client's incoming connections.
 * Stripping Headers, Path, ContentType and Body, then
 * places them into the Request type provided by the
 * package.
 */
func parseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)
	dat, err := reader.ReadString('\n')
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return &Request{}, fmt.Errorf("Timeout reading request line: %w", err)
		}
		return &Request{}, err
	}

	log.Printf("Header: %s\n", dat)

	part := strings.Fields(strings.TrimSpace(dat))
	if len(part) != 3 {
		return &Request{}, fmt.Errorf("Invalid request format: %s. Usage: <Method> <Path> <metadata>", dat)
	}
	var protVer string
	if strings.Contains(part[2], "HTTP/") {
		protVer = part[2]
	}

	header := RequestHeader{
		Method:          part[0],
		Path:            part[1],
		ProtocolVersion: protVer,
		ContentLength:   0,
	}

	var cntLen int
	// Read connection payload by line until end of file
	for {
		dat, err = reader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return &Request{}, fmt.Errorf("Timeout reading headers: %w", err)
			}
			return &Request{}, err
		}
		dat = strings.TrimSpace(dat)
		if dat == "" {
			// End of headers / No content-type provided
			break
		}

		if strings.HasPrefix(strings.ToLower(dat), "content-type") {
			ctnType := strings.SplitN(dat, ":", 2)
			if len(ctnType) == 2 {
				header.ContentType = strings.TrimSpace(ctnType[1])
			}
		}

		if strings.HasPrefix(dat, "Content-Length") {
			cl := strings.SplitN(dat, ":", 2)
			if len(cl) == 2 {
				cntLen, err = strconv.Atoi(strings.TrimSpace(cl[1]))
				if err != nil {
					return &Request{}, fmt.Errorf("Invalid Content-Length: %s", cl[1])
				}
			}
		}
	}

	// Clear the read deadline now that headers are parsed successfully
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return &Request{}, fmt.Errorf("Failed to clear read deadline: %w", err)
	}

	// Copy body into type request body field
	bodyBytes := new(bytes.Buffer)
	if cntLen > 0 {
		header.ContentLength = cntLen
	}

	_, err = copyBody(bodyBytes, reader, header.ContentLength)
	if err != nil && err != io.EOF {
		return &Request{}, fmt.Errorf("Failed to read request body: %w", err)
	}

	return &Request{Header: header, Body: bodyBytes, params: make(map[string]string)}, nil
}

func copyBody(dst *bytes.Buffer, src *bufio.Reader, cntLen int) (n int64, err error) {
	return io.CopyN(dst, src, int64(cntLen))
}
