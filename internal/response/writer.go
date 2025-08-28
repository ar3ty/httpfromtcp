package response

import (
	"fmt"
	"io"

	"github.com/ar3ty/httpfromtcp/internal/headers"
)

type WriterState int

const (
	writingStatusLine WriterState = iota
	writingHeaders
	writingBody
	writingTrailers
)

type Writer struct {
	state  WriterState
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		state:  writingStatusLine,
		writer: w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writingStatusLine {
		return fmt.Errorf("writing status-line is not allowed in current state")
	}
	defer func() { w.state = writingHeaders }()

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

	_, err := w.writer.Write(status)
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writingHeaders {
		return fmt.Errorf("writing headers is not allowed in current state")
	}
	defer func() { w.state = writingBody }()

	for key, value := range headers {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writingBody {
		return 0, fmt.Errorf("writing body is not allowed in current state")
	}
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != writingBody {
		return 0, fmt.Errorf("writing body is not allowed in current state")
	}

	total := 0
	num := []byte(fmt.Sprintf("%x\r\n", len(p)))
	n, err := w.writer.Write(num)
	if err != nil {
		return n, err
	}
	total += n

	n, err = w.writer.Write(p)
	if err != nil {
		return total, err
	}
	total += n

	n, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return total, err
	}
	total += n

	return total, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != writingBody {
		return 0, fmt.Errorf("writing body is not allowed in current state")
	}
	defer func() { w.state = writingTrailers }()

	num := []byte("0\r\n")
	n, err := w.writer.Write(num)
	if err != nil {
		return n, err
	}
	return 0, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writingTrailers {
		return fmt.Errorf("writing trailers is not allowed in current state")
	}
	defer func() { w.state = writingBody }()

	for key, value := range h {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	return err
}
