package response

import (
	"io"
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
		status = ""
	}
	w.Write([]byte(status))
}
