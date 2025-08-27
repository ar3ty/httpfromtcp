package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/ar3ty/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	portStr := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", portStr)
	if err != nil {
		return nil, fmt.Errorf("couldn't create listener: %v", err)
	}
	server := &Server{
		listener: listener,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Couldn't get connection: %v", err)
			continue
		}
		fmt.Printf("Connection accepted from: %s\n", conn.RemoteAddr())

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	headers := response.GetDefaultHeaders(0)
	err := response.WriteStatusLine(conn, response.OK)
	if err != nil {
		log.Fatalf("Error writing in connection: %v", err)
	}
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Fatalf("Error writing in connection: %v", err)
	}
}
