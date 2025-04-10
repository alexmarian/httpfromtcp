package response

import (
	"fmt"
	"github.com/alexmarian/httpfromtcp/internal/headers"
	"io"
	"os"
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
type writerState int

const (
	writerStateInitialized writerState = iota
	writerStateResponseLineWrote
	writerStateHeadersWrote
	writerStateDone
)

type Writer struct {
	io.Writer
	writerState writerState
}

func (he HandlerError) Write(w *Writer) {
	switch he.StatusCode {
	case BAD_REQUEST:
		w.WriteFile("html/bad_request.html", "text/html", BAD_REQUEST)
	case INTERNAL_SERVER_ERROR:
		w.WriteFile("html/internal_server_error.html", "text/html", INTERNAL_SERVER_ERROR)
	default:
		w.WriteStatusLine(he.StatusCode)
		messageBytes := []byte(he.Message)
		headers := GetDefaultHeaders(len(messageBytes))
		w.WriteHeaders(headers)
		w.WriteBody(messageBytes)
	}

}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer:      w,
		writerState: writerStateInitialized,
	}
}

func (w *Writer) WriteFile(file, contentType string, code StatusCode) (int, *HandlerError) {
	fileContent, err := os.ReadFile(file)
	if err != nil {
		return 0, &HandlerError{
			StatusCode: INTERNAL_SERVER_ERROR,
			Message:    fmt.Sprintf("Failed to load %s", file),
		}
	}
	w.WriteStatusLine(code)
	defaultHeaders := GetDefaultHeaders(len(fileContent))
	defaultHeaders.Override(headers.ContentTypeHeader, contentType)
	w.WriteHeaders(defaultHeaders)
	w.WriteBody(fileContent)
	return len(fileContent), nil
}
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateInitialized {
		return fmt.Errorf("wrong state: %d, expected: %d", w.writerState, writerStateInitialized)
	}
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
	w.writerState = writerStateResponseLineWrote
	return nil
}
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateResponseLineWrote {
		return fmt.Errorf("wrong state: %d, expected: %d", w.writerState, writerStateHeadersWrote)
	}
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
	w.writerState = writerStateHeadersWrote
	return nil
}
func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateHeadersWrote {
		return 0, fmt.Errorf("wrong state: %d, expected: %d", w.writerState, writerStateHeadersWrote)
	}
	w.Write(p)
	w.Write([]byte("\n"))
	w.writerState = writerStateDone
	return len(p), nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	header := headers.NewHeaders()
	header.Set(headers.ContentTypeHeader, "text/plain")
	header.Set(headers.ConnectionHeader, "close")
	header.Set(headers.ContentLengthHeader, fmt.Sprintf("%d", contentLen))
	return header
}
