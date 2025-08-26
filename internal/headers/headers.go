package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

var tokenChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

func checkFieldName(name []byte) bool {
	for _, char := range name {
		if char >= 'A' && char <= 'Z' ||
			char >= 'a' && char <= 'z' ||
			char >= '0' && char <= '9' {
			continue
		}
		if slices.Contains(tokenChars, char) {
			continue
		}
		return false
	}
	return true
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	if val, ok := h[key]; ok {
		value = val + ", " + value
	}
	h[key] = value
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}

	parts := bytes.SplitN(data[:idx], []byte(":"), 2)
	key := string(parts[0])
	if key != strings.TrimRight(key, " ") {
		return 0, false, fmt.Errorf("invalid field line format: %s", key)
	}
	key = strings.TrimSpace(key)

	if !checkFieldName([]byte(key)) {
		return 0, false, fmt.Errorf("forbidden symbol in header: %s", key)
	}

	value := string(bytes.TrimSpace(parts[1]))

	h.Set(key, value)
	return idx + 2, false, nil
}
