package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
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
		case "/video":
			w.WriteFile("assets/vim.mp4", "video/mp4", response.SUCCESS)

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
	hs := headers.NewHeaders()
	hs.Set(headers.ContentTypeHeader, resp.Header.Get(headers.ContentTypeHeader))
	hs.Set(headers.TransferEncodingHeader, "chunked")
	hs.Set(headers.TrailerHeader, strings.Join([]string{headers.XContentSHA256Trailer, headers.XContentSLengthTrailer}, ", "))
	w.WriteHeaders(hs)
	buffer := make([]byte, 1024)
	b := bytes.Buffer{}
	body := resp.Body
	for {
		br, err := body.Read(buffer)
		if br > 0 {
			b.Write(buffer[:br])
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
			trailers := headers.NewHeaders()
			allResponseBytes := b.Bytes()
			trailers.Set(headers.XContentSHA256Trailer, fmt.Sprintf("%x", sha256.Sum256(allResponseBytes)))
			trailers.Set(headers.XContentSLengthTrailer, fmt.Sprintf("%d", len(allResponseBytes)))
			w.WriteTrailers(trailers)
			return nil
		}
	}
	return nil
}
