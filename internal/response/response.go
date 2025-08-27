package response

import (
	"fmt"
	"io"

	"github.com/ar3ty/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var status string
	switch statusCode {
	case OK:
		status = "HTTP/1.1 200 OK"
	case BadRequest:
		status = "HTTP/1.1 400 Bad Request"
	case InternalServerError:
		status = "HTTP/1.1 500 Internal Server Error"
	default:
		status = fmt.Sprintf("HTTP/1.1 %d ", statusCode)
	}
	status += "\r\n"

	_, err := w.Write([]byte(status))
	if err != nil {
		return fmt.Errorf("failed writing status-line: %w", err)
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	data := ""
	for key, value := range headers {
		data += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	data += "\r\n"

	_, err := w.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("failed writing headers: %w", err)
	}
	return nil
}
