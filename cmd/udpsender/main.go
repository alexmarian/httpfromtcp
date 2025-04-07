package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"syscall"
)

const port = 42069

func main() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		fmt.Printf("Error resolving address: %v\n", err)
		syscall.Exit(-1)
	}
	udp, err := net.DialUDP("udp", nil, addr)
	defer udp.Close()
	if err != nil {
		fmt.Printf("Error dialing UDP: %v\n", err)
		syscall.Exit(-1)
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">")
		line, _, err := reader.ReadLine()
		if err != nil {
			fmt.Printf("Error reading line: %v\n", err)
			break
		}
		_, err = udp.Write(line)
		if err != nil {
			fmt.Printf("Error writing line: %v\n", err)
			break
		}
	}

}
