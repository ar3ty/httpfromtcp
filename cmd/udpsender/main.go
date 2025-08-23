package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const serveAddr = "localhost:42069"

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", serveAddr)
	if err != nil {
		log.Fatalf("Cannot resolve UDP address: %s", err)
	}
	connUDP, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Cannot open UDP connection: %s", err)
	}
	defer connUDP.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failure reading string: %s\n", err.Error())
		}
		_, err = connUDP.Write([]byte(line))
		if err != nil {
			log.Fatalf("Failure sending message: %s\n", err.Error())
		}

		fmt.Printf("Message sent: %s", line)
	}
}
