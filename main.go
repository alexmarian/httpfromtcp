package main

import (
	"fmt"
	"github.com/alexmarian/httpfromtcp/internal/utils"
	"log"
	"os"
)

const inputFilePath = "messages.txt"

func main() {
	f, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("could not open %s: %s\n", inputFilePath, err)
	}

	fmt.Printf("Reading data from %s\n", inputFilePath)
	fmt.Println("=====================================")

	lines := utils.GetLinesChannel(f)
	for line := range lines {
		fmt.Printf("read: %s", line)
	}
}
