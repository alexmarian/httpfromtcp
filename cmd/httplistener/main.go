package main

import (
	"github.com/alexmarian/httpfromtcp/internal/request"
	"github.com/alexmarian/httpfromtcp/internal/response"
	"github.com/alexmarian/httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
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

func handler(w response.Writer, req *request.Request) *response.HandlerError {
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
	return nil
}
