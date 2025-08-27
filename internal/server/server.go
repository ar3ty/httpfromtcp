package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/ar3ty/httpfromtcp/internal/request"
	"github.com/ar3ty/httpfromtcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    []byte
}

func (e *HandlerError) report(w io.Writer) {
	headers := response.GetDefaultHeaders(len(e.Message))

	err := response.WriteStatusLine(w, e.StatusCode)
	if err != nil {
		log.Fatalf("Error writing in connection: %v", err)
	}
	err = response.WriteHeaders(w, headers)
	if err != nil {
		log.Fatalf("Error writing in connection: %v", err)
	}

	_, err = w.Write(e.Message)
	if err != nil {
		log.Fatalf("Error writing in connection: %v", err)
	}
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		errToReport := &HandlerError{
			StatusCode: response.BadRequest,
			Message:    []byte(err.Error()),
		}
		errToReport.report(conn)
		return
	}
	buf := bytes.NewBuffer([]byte{})

	handlerErr := s.handler(buf, req)
	if handlerErr != nil {
		handlerErr.report(conn)
		return
	}

	headers := response.GetDefaultHeaders(buf.Len())

	err = response.WriteStatusLine(conn, response.OK)
	if err != nil {
		log.Fatalf("Error writing in connection: %v", err)
	}
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Fatalf("Error writing in connection: %v", err)
	}

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		log.Fatalf("Error writing in connection: %v", err)
	}
}
