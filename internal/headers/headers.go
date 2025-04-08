package headers

import (
	"bytes"
	"fmt"
	"strings"
)

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

	(*headers).Set(strings.TrimSpace(name), strings.TrimSpace(string(parts[1])))
	return nil

}

func (h Headers) Set(key, value string) {
	h[key] = value
}
