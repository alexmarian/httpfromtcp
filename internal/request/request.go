package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {
	requestI, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	firstSeparator := bytes.Index(requestI, []byte(crlf))

	line, err := parseRequestLine(string(requestI[:firstSeparator]))
	if err != nil {
		return nil, err
	}
	request := Request{RequestLine: line}
	return &request, nil
}

func parseRequestLine(requestLine string) (RequestLine, error) {
	requestLineParts := strings.Split(requestLine, " ")
	if len(requestLineParts) != 3 {
		return RequestLine{}, errors.New("invalid request line")
	}
	method := requestLineParts[0]
	if !isAllUppercase(method) {
		return RequestLine{}, errors.New("invalid method")
	}
	target := requestLineParts[1]
	if requestLineParts[2] != "HTTP/1.1" {
		return RequestLine{}, errors.New("invalid HTTP version")
	}
	version := strings.Split(requestLineParts[2], "/")[1]
	return RequestLine{
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
