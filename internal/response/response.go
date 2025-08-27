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
	var reasonPhrase string
	switch statusCode {
	case OK:
		reasonPhrase = "OK"
	case BadRequest:
		reasonPhrase = "Bad Request"
	case InternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		reasonPhrase = ""
	}
	status := []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase))

	_, err := w.Write(status)
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
