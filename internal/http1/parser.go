package http1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ReadRequest parses a basic HTTP request from a bufio.Reader.
func ReadRequest(r *bufio.Reader) (*http.Request, error) {
	// Read and parse the Request Line: e.g., "GET /index.html HTTP/1.1\r\n"
	var line string
	var err error
	for {
		line, err = readCRLFLine(r)
		if err != nil {
			return nil, fmt.Errorf("reading request line: %w", err)
		}
		// Ignore any empty lines (CRLF) received prior to the request-line (RFC 9112 §2.2)
		if line != "" {
			break
		}
	}

	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, errors.New("malformed HTTP request line")
	}

	method, rawURI, proto := parts[0], parts[1], parts[2]

	parsedURL, err := url.ParseRequestURI(rawURI)
	if err != nil {
		return nil, fmt.Errorf("invalid request target URI: %w", err)
	}

	// Read headers until an empty line is reached
	headers := make(http.Header)
	for {
		headerLine, err := readCRLFLine(r)
		if err != nil {
			return nil, fmt.Errorf("reading headers: %w", err)
		}
		if headerLine == "" {
			// An empty line signals the end of HTTP headers
			break
		}

		key, value, found := strings.Cut(headerLine, ":")
		if !found {
			return nil, fmt.Errorf("malformed header line: %q", headerLine)
		}

		// Header values can have leading/trailing whitespace
		headers.Add(strings.TrimSpace(key), strings.TrimSpace(value))
	}

	// Setup Body based on Content-Length (if present)
	var body io.ReadCloser = http.NoBody
	var contentLength int64 = 0

	if clStr := headers.Get("Content-Length"); clStr != "" {
		cl, err := strconv.ParseInt(clStr, 10, 64)
		if err != nil || cl < 0 {
			return nil, fmt.Errorf("invalid Content-Length: %q", clStr)
		}
		contentLength = cl
		// io.LimitReader ensures we only read the declared body bytes
		body = io.NopCloser(io.LimitReader(r, cl))
	}

	protoMajor, protoMinor, ok := http.ParseHTTPVersion(proto)
	if !ok {
		return nil, fmt.Errorf("malformed HTTP version: %q", proto)
	}

	req := &http.Request{
		Method:        method,
		URL:           parsedURL,
		Proto:         proto,
		ProtoMajor:    protoMajor,
		ProtoMinor:    protoMinor,
		Header:        headers,
		Body:          body,
		ContentLength: contentLength,
		Host:          headers.Get("Host"),
	}

	return req, nil
}

// readCRLFLine reads bytes up to \n and strips trailing \r\n or \n.
func readCRLFLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Strip \r\n or trailing \n
	line = strings.TrimRight(line, "\r\n")
	return line, nil
}
