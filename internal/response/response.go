package response

import (
	"fmt"
	"github.com/alexmarian/httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

const (
	SUCCESS               StatusCode = 200
	BAD_REQUEST           StatusCode = 400
	INTERNAL_SERVER_ERROR StatusCode = 500
)

type HandlerError struct {
	StatusCode StatusCode
	Message    string
}

func (he HandlerError) Write(w io.Writer) {
	WriteStatusLine(w, he.StatusCode)
	messageBytes := []byte(he.Message)
	headers := GetDefaultHeaders(len(messageBytes))
	WriteHeaders(w, headers)
	w.Write(messageBytes)
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case SUCCESS:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return err
		}
	case BAD_REQUEST:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return err
		}
	case INTERNAL_SERVER_ERROR:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported status code: %d", statusCode)
	}
	return nil
}
func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for name, value := range headers {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", name, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	header := headers.NewHeaders()
	header.Set(headers.ContentTypeHeader, "text/plain")
	header.Set(headers.ConnectionHeader, "close")
	header.Set(headers.ContentLengthHeader, fmt.Sprintf("%d", contentLen))
	return header
}
