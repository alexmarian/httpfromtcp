package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

const ContentLengthHeader = "Content-Length"
const ContentTypeHeader = "Content-Type"
const ConnectionHeader = "Connection"

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

const crlf = "\r\n"
const column = ":"

func (h *Headers) Parse(data []byte) (n int, done bool, err error) {
	index := bytes.Index(data, []byte(crlf))
	if index == -1 {
		return 0, false, nil
	}
	if index == 0 {
		return 2, true, nil
	}

	if err := parseHeaderLine(data[:index], h); err != nil {
		return 0, false, fmt.Errorf("invalid header line: %w", err)
	}

	return index + 2, false, nil
}

func parseHeaderLine(data []byte, headers *Headers) error {
	parts := bytes.SplitN(data, []byte(":"), 2)
	name := string(parts[0])

	if name == "" || strings.HasSuffix(name, " ") {
		return fmt.Errorf("invalid header line name: %s", name)
	}

	name = strings.TrimSpace(name)
	if !isValidName(name) {
		return fmt.Errorf("invalid header line name: %s", name)
	}
	(*headers).Set(name, strings.TrimSpace(string(parts[1])))
	return nil

}

func isValidName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for _, r := range name {
		if !isValidHeaderNameChar(byte(r)) {
			return false
		}
	}
	return true
}

func isValidHeaderNameChar(r byte) bool {
	specials := []byte{
		'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~',
	}
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || slices.Contains(specials, r)
}

func (h Headers) Set(key, value string) {
	evalue, present := h[strings.ToLower(key)]
	if present {
		h[strings.ToLower(key)] = fmt.Sprintf("%s, %s", evalue, value)
	} else {
		h[strings.ToLower(key)] = value
	}

}
func (h Headers) Override(key, value string) {
	h[strings.ToLower(key)] = value
}

func (h Headers) Get(key string) (string, bool) {
	val, present := h[strings.ToLower(key)]
	return val, present
}
