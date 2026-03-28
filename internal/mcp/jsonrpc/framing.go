package jsonrpc

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// readMessage reads a JSON-RPC message, auto-detecting between
// Content-Length framing (LSP-style) and bare JSON lines (newline-delimited).
func readMessage(reader *bufio.Reader) ([]byte, error) {
	for {
		// Peek at the first byte to decide framing style
		firstBytes, err := reader.Peek(1)
		if err != nil {
			return nil, err
		}

		if firstBytes[0] == '{' {
			// Bare JSON line
			line, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				return nil, err
			}
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			return line, nil
		}

		if firstBytes[0] == '\n' || firstBytes[0] == '\r' {
			// Skip blank lines between messages
			_, _ = reader.ReadByte()
			continue
		}

		// Content-Length framed message
		return readContentLengthMessage(reader)
	}
}

func readContentLengthMessage(reader *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if line == "\r\n" || line == "\n" {
			break
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(key), "Content-Length") {
			parsed, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("parse content length %q: %w", strings.TrimSpace(value), err)
			}
			contentLength = parsed
		}
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// writeMessage writes a JSON-RPC message as a bare JSON line (newline-delimited).
// This is compatible with both bare-JSON-line servers and Content-Length servers
// that also accept newline-delimited input.
func writeMessage(writer io.Writer, payload []byte) error {
	_, err := writer.Write(append(payload, '\n'))
	return err
}
