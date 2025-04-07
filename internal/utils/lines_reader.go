package utils

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

func GetLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go readLines(f, lines)
	return lines
}

func readLines(f io.ReadCloser, lines chan<- string) {
	defer f.Close()
	defer close(lines)
	currentLineContents := ""
	for {
		buffer := make([]byte, 8, 8)
		n, err := f.Read(buffer)
		if err != nil {
			if currentLineContents != "" {
				lines <- currentLineContents
				currentLineContents = ""
			}
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Printf("error: %s\n", err.Error())
			break
		}
		str := string(buffer[:n])
		parts := strings.Split(str, "\n")
		for i := 0; i < len(parts)-1; i++ {
			lines <- fmt.Sprintf("%s%s", currentLineContents, parts[i])
			currentLineContents = ""
		}
		currentLineContents += parts[len(parts)-1]
	}
}
