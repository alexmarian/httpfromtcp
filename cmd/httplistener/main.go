package main

import (
	"fmt"
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
		return proxyHandler(w, req)
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
func proxyHandler(w *response.Writer, req *request.Request) *response.HandlerError {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbin.org/" + target
	fmt.Println("Proxying to", url)
	resp, err := http.Get(url)
	if err != nil {
		return &response.HandlerError{
			StatusCode: response.INTERNAL_SERVER_ERROR,
			Message:    "Error fetching data from httpbin",
		}
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.SUCCESS)
	h := response.GetDefaultHeaders(0)
	h.Override("Transfer-Encoding", "chunked")
	h.Remove("Content-Length")
	w.WriteHeaders(h)

	const maxChunkSize = 1024
	buffer := make([]byte, maxChunkSize)
	for {
		n, err := resp.Body.Read(buffer)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}
	return nil
}
