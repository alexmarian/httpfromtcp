package main

import (
	"github.com/alexmarian/httpfromtcp/internal/headers"
	"github.com/alexmarian/httpfromtcp/internal/request"
	"github.com/alexmarian/httpfromtcp/internal/response"
	"github.com/alexmarian/httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) *response.HandlerError {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		handlerError := handleProxy(w, req)
		if handlerError != nil {
			return handlerError
		}
	} else {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &response.HandlerError{
				StatusCode: response.BAD_REQUEST,
				Message:    "Your problem is not my problem",
			}
		case "/myproblem":
			return &response.HandlerError{
				StatusCode: response.INTERNAL_SERVER_ERROR,
				Message:    "Woopsie, my bad",
			}
		default:
			w.WriteFile("html/success.html", "text/html", response.SUCCESS)
		}
	}
	return nil
}

func handleProxy(w *response.Writer, req *request.Request) *response.HandlerError {
	location := "https://httpbin.org/" + strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	resp, err := http.Get(location)
	defer resp.Body.Close()
	if err != nil {
		return &response.HandlerError{
			StatusCode: response.INTERNAL_SERVER_ERROR,
			Message:    "Error fetching data from httpbin",
		}
	}
	w.WriteStatusLine(response.SUCCESS)
	hs := response.GetDefaultHeaders(0)
	hs.Remove(headers.ContentLengthHeader)
	hs.Set(headers.ContentTypeHeader, resp.Header.Get(headers.ContentTypeHeader))
	hs.Set(headers.TransferEncodingHeader, "chunked")
	w.WriteHeaders(hs)
	buffer := make([]byte, 1024)
	body := resp.Body
	for {
		br, err := body.Read(buffer)
		if br > 0 {
			_, err := w.WriteChunkedBody(buffer[:br])
			if err != nil {
				log.Println("Error writing chunked body:", err)
				return &response.HandlerError{
					StatusCode: response.INTERNAL_SERVER_ERROR,
					Message:    err.Error(),
				}
			}
		}
		if err != nil && err == io.EOF {
			w.WriteChunkedBodyDone()
			return nil
		}
	}
	return nil
}
