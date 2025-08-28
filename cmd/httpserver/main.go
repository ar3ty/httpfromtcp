package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ar3ty/httpfromtcp/internal/headers"
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

func handlerVideo(w *response.Writer, req *request.Request) {
	message, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		handler500(w, req)
		return
	}

	w.WriteStatusLine(response.OK)
	h := response.GetDefaultHeaders(len(message))
	h.Replace("Content-Type", "video/mp4")
	w.WriteHeaders(h)
	w.WriteBody(message)
}

func handlerProxy(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	url := "https://httpbin.org" + target
	fmt.Println("Proxying to", url)

	resp, err := http.Get(url)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.OK)

	h := response.GetDefaultHeaders(0)
	h.Delete("Content-Length")
	h.Replace("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	w.WriteHeaders(h)

	buf := make([]byte, 1024)
	fullBody := []byte{}

	for {
		n, err := resp.Body.Read(buf)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			_, err := w.WriteChunkedBody(buf[:n])
			if err != nil {
				fmt.Println("Error writing chunk body:", err)
			}
			fullBody = append(fullBody, buf[:n]...)

		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}

	}
	hash := sha256.Sum256(fullBody)

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunk body done:", err)
	}

	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hash))
	trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("Error writing trailers:", err)
	}
	fmt.Println("Wrote trailers")
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
		handlerProxy(w, req)
		return
	}
	if target == "/video" {
		handlerVideo(w, req)
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
