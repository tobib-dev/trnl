package trnl

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

	part := strings.SplitN(strings.TrimSpace(dat), " ", 3)
	if len(part) != 3 {
		return &Request{}, fmt.Errorf("Invalid request format: %s. Usage: <Method> <Path> <Body>", dat)
	}

	header := RequestHeader{
		Method: part[0],
		Path:   part[1],
	}

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
