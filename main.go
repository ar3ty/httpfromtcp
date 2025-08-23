package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilePath = "messages.txt"

func main() {
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("Cannot open the %s: %s", inputFilePath, err)
	}

	linesChannel := getLinesChannel(file)
	for line := range linesChannel {
		fmt.Println("read:", line)
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	linesChannel := make(chan string)
	go func() {
		currentLineBuffer := ""
		defer f.Close()
		defer close(linesChannel)
		for {
			eight := make([]byte, 8)
			n, err := f.Read(eight)
			if err != nil {
				if errors.Is(err, io.EOF) {
					linesChannel <- currentLineBuffer
					currentLineBuffer = ""
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			line := string(eight[:n])
			parts := strings.Split(line, "\n")
			for i := 0; i < len(parts)-1; i++ {
				linesChannel <- currentLineBuffer + parts[i]
				currentLineBuffer = ""
			}
			currentLineBuffer += parts[len(parts)-1]
		}
	}()
	return linesChannel
}
