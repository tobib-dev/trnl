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
		return &Request{}, err
	}
	// Log request header and time
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
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
	}

	var cntLen int
	// Read connection payload by line until end of file
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
					return &Request{}, fmt.Errorf("invalid Content-Length: %s", cl[1])
				}
			}
		}
	}

	// Copy body into type request body field
	bodyBytes := new(bytes.Buffer)
	if cntLen > 0 {
		header.ContentLength = cntLen
	} else {
		header.ContentLength = 0
	}
	n, err := copyBody(bodyBytes, reader, header.ContentLength)
	if err != nil && err != io.EOF {
		return &Request{}, fmt.Errorf("failed to read request body: %w", err)
	}
	if n != int64(header.ContentLength) {
		return &Request{}, fmt.Errorf("incomplete request body: read %d bytes, expected %d", n, cntLen)
	}

	return &Request{Header: header, Body: bodyBytes}, nil
}

func copyBody(dst *bytes.Buffer, src *bufio.Reader, cntLen int) (n int64, err error) {
	return io.CopyN(dst, src, int64(cntLen))
}
