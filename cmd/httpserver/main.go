package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ar3ty/httpfromtcp/internal/request"
	"github.com/ar3ty/httpfromtcp/internal/response"
	"github.com/ar3ty/httpfromtcp/internal/server"
)

const port = 42069

func handler400(w *response.Writer, _ *request.Request) {
	message := []byte("<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>")

	w.WriteStatusLine(response.BadRequest)
	h := response.GetDefaultHeaders(len(message))
	h.Replace("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(message)
}

func handler500(w *response.Writer, _ *request.Request) {
	message := []byte("<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>")

	w.WriteStatusLine(response.InternalServerError)
	h := response.GetDefaultHeaders(len(message))
	h.Replace("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(message)
}

func handler200(w *response.Writer, _ *request.Request) {
	message := []byte("<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>")

	w.WriteStatusLine(response.OK)
	h := response.GetDefaultHeaders(len(message))
	h.Replace("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(message)
}

func handlerHTTPBin(w *response.Writer, req *request.Request, query string) {
	w.WriteStatusLine(response.OK)

	h := response.GetDefaultHeaders(0)
	h.Delete("Content-Length")
	h.Set("Transfer-Encoding", "chunked")
	w.WriteHeaders(h)

	resp, err := http.Get("https://httpbin.org" + query)
	if err != nil {
		handler500(w, req)
		return
	}

	buf := make([]byte, 1024)

	for {
		n, err := resp.Body.Read(buf)
		if err != nil {
			break
		}

		w.WriteChunkedBody(buf[:n])
	}
	w.WriteChunkedBodyDone()
}

func handler(w *response.Writer, req *request.Request) {
	target := req.RequestLine.RequestTarget
	if target == "/yourproblem" {
		handler400(w, req)
		return
	}
	if target == "/myproblem" {
		handler500(w, req)
		return
	}
	if strings.HasPrefix(target, "/httpbin") {
		handlerHTTPBin(w, req, strings.TrimPrefix(target, "/httpbin"))
		return
	}

	handler200(w, req)
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
