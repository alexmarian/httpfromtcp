package main

import (
	"fmt"
	"github.com/alexmarian/httpfromtcp/internal/utils"
	"log"
	"net"
)

const inputFilePath = "messages.txt"
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
		connectionChannel := utils.GetLinesChannel(accept)
		for line := range connectionChannel {
			fmt.Printf("%s\n", line)
		}

	}
}
