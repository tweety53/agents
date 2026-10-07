package lspmux

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// readMessage reads one LSP base-protocol message — header lines, a blank
// line, then Content-Length bytes — and returns its body. io.EOF means the
// stream ended cleanly between messages; a stream cut inside one is
// io.ErrUnexpectedEOF.
func readMessage(r *bufio.Reader) ([]byte, error) {
	length, headers := -1, 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF && line == "" && headers == 0 {
				return nil, io.EOF
			}
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		headers++
		name, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || n < 0 {
				return nil, fmt.Errorf("lspmux: bad Content-Length %q", value)
			}
			length = n
		}
	}
	if length < 0 {
		return nil, errors.New("lspmux: message without a Content-Length header")
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

// writeMessage frames body and writes it in one Write. Callers sharing w
// serialise their calls.
func writeMessage(w io.Writer, body []byte) error {
	_, err := w.Write(append([]byte(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))), body...))
	return err
}
