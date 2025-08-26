package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ar3ty/httpfromtcp/internal/headers"
)

type parseStatus int

const (
	parseStatusInitialized parseStatus = iota
	parseStatusParsingHeaders
	parseStatusDone
)
const bufferSize = 8

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Status      parseStatus
}

func requestLineFromString(line string) (*RequestLine, error) {
	methods := map[string]struct{}{
		"GET":     {},
		"HEAD":    {},
		"POST":    {},
		"PUT":     {},
		"DELETE":  {},
		"CONNECT": {},
		"OPTIONS": {},
		"TRACE":   {},
	}

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, errors.New("invalid request line structure")
	}
	method := parts[0]
	if _, ok := methods[method]; !ok {
		return nil, errors.New("invalid method")
	}
	target := parts[1]
	if len(target) == 0 {
		return nil, errors.New("invalid target")
	}
	protocolVersion := strings.Split(parts[2], "/")
	if len(protocolVersion) != 2 {
		return nil, errors.New("invalid http version variable")
	}
	if protocolVersion[0] != "HTTP" {
		return nil, fmt.Errorf("invalid http version variable: %s", protocolVersion[0])
	}
	if protocolVersion[1] != "1.1" {
		return nil, fmt.Errorf("invalid http version variable: %s", protocolVersion[1])
	}
	return &RequestLine{
		HttpVersion:   protocolVersion[1],
		RequestTarget: target,
		Method:        method,
	}, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, idx, err
	}
	return requestLine, idx + 2, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.Status {
	case parseStatusInitialized:
		rLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *rLine
		r.Status = parseStatusParsingHeaders
		return n, nil
	case parseStatusParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.Status = parseStatusDone
		}
		return n, nil
	case parseStatusDone:
		return 0, errors.New("trying to read data in a done state")
	default:
		return 0, errors.New("unknown status")
	}
}

func (r *Request) parse(data []byte) (int, error) {
	totalParsed := 0
	for r.Status != parseStatusDone {
		n, err := r.parseSingle(data[totalParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return totalParsed, nil
		}
		totalParsed += n
	}
	return totalParsed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		Headers: headers.Headers{},
		Status:  parseStatusInitialized,
	}

	for req.Status != parseStatusDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, 2*len(buf))
			_ = copy(newBuf, buf)
			buf = newBuf
		}
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				if req.Status != parseStatusDone {
					return nil, fmt.Errorf("incomplete request")
				}
				break
			}
			return nil, err
		}
		readToIndex += n

		n, err = req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if n != 0 {
			length := readToIndex - n
			_ = copy(buf, buf[n:readToIndex])
			readToIndex = length
		}
	}

	return req, nil
}
