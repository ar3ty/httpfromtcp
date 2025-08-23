package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Cannot open the tcp connection on port '%s': %s", port, err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Cannot get connection: %s\n", err)
			continue
		}
		fmt.Printf("Connection accepted from: %s\n", conn.RemoteAddr())

		linesChannel := getLinesChannel(conn)

		for line := range linesChannel {
			fmt.Println(line)
		}
		fmt.Printf("Connection to %s is closed\n", conn.RemoteAddr())
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	linesChannel := make(chan string)
	go func() {
		defer f.Close()
		defer close(linesChannel)

		currentLineBuffer := ""
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
