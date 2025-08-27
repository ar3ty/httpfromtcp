package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ar3ty/httpfromtcp/internal/request"
	"github.com/ar3ty/httpfromtcp/internal/response"
	"github.com/ar3ty/httpfromtcp/internal/server"
)

const port = 42069

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: response.BadRequest,
			Message:    []byte("Your problem is not my problem\n"),
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: response.InternalServerError,
			Message:    []byte("Woopsie, my bad\n"),
		}
	default:
		message := "All good, frfr\n"
		_, err := w.Write([]byte(message))
		if err != nil {
			return &server.HandlerError{
				StatusCode: response.InternalServerError,
				Message:    []byte("Woopsie, my bad\n"),
			}
		}
	}
	return nil
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
