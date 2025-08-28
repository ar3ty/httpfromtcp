package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/ar3ty/httpfromtcp/internal/request"
	"github.com/ar3ty/httpfromtcp/internal/response"
)

func report(w *response.Writer, code response.StatusCode, messagestr string) {
	err := w.WriteStatusLine(code)
	if err != nil {
		log.Fatalf("Error writing in connection: %v", err)
	}
	message := []byte(messagestr)

	err = w.WriteHeaders(response.GetDefaultHeaders(len(message)))
	if err != nil {
		log.Fatalf("Error writing in connection: %v", err)
	}

	if len(message) > 0 {
		_, err = w.WriteBody(message)
		if err != nil {
			log.Fatalf("Error writing in connection: %v", err)
		}
	}
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	portStr := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", portStr)
	if err != nil {
		return nil, fmt.Errorf("couldn't create listener: %v", err)
	}
	server := &Server{
		listener: listener,
		handler:  handler,
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

	resWriter := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		report(resWriter, 500, "couldn't get request")
		return
	}

	s.handler(resWriter, req)
}
