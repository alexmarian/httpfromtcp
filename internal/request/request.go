package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateDone
)

type Request struct {
	RequestLine RequestLine

	state requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	req := &Request{
		state: requestStateInitialized,
	}
	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(io.EOF, err) {
				req.state = requestStateDone
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}
	return req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + 2, nil
}

func requestLineFromString(requestLine string) (*RequestLine, error) {
	requestLineParts := strings.Split(requestLine, " ")
	if len(requestLineParts) != 3 {
		return nil, errors.New("invalid request line")
	}
	method := requestLineParts[0]
	if !isAllUppercase(method) {
		return nil, errors.New("invalid method")
	}
	target := requestLineParts[1]
	if requestLineParts[2] != "HTTP/1.1" {
		return nil, errors.New("invalid HTTP version")
	}
	version := strings.Split(requestLineParts[2], "/")[1]
	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   version,
	}, nil
}

func isAllUppercase(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) || !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			// something actually went wrong
			return 0, err
		}
		if n == 0 {
			// just need more data
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateDone
		return n, nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}
