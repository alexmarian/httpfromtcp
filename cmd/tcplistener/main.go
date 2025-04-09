package main

import (
	"fmt"
	"github.com/alexmarian/httpfromtcp/internal/request"
	"log"
	"net"
)

const port = 42069

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Error listening on port %d: %v", port, err)
	}

	defer listener.Close()
	for {
		accept, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connections: %v", err)
		}
		log.Printf("Accepted connection: %v\n", accept)
		req, err := request.RequestFromReader(accept)
		if err != nil {
			log.Fatalf("Error reading request: %v", err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for name, value := range req.Headers {
			fmt.Printf("- %s: %s\n", name, value)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))
	}
}
