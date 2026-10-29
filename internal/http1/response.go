package http1

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
)

type Response struct {
	Status      string
	StatusCode  int
	HeaderMap   http.Header
	Request     *http.Request
	conn        net.Conn
	wroteHeader bool
	chunked     bool
}

func NewResponse(conn net.Conn) *Response {
	return &Response{
		HeaderMap: make(http.Header),
		conn:      conn,
	}
}

func (r *Response) Header() http.Header {
	if r.HeaderMap == nil {
		r.HeaderMap = make(http.Header)
	}
	return r.HeaderMap
}

func (r *Response) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}

	r.StatusCode = statusCode
	r.Status = http.StatusText(statusCode)
	if r.Status == "" {
		r.Status = "status code " + strconv.Itoa(statusCode)
	}

	// Determine framing: Content-Length vs Chunked
	hasContentLength := r.HeaderMap.Get("Content-Length") != ""
	// Responses like 204 No Content, 304 Not Modified, or 1xx cannot have a body
	bodyAllowed := statusCode >= 200 && statusCode != http.StatusNoContent && statusCode != http.StatusNotModified

	if bodyAllowed && !hasContentLength {
		// Use chunked encoding if no Content-Length was provided
		r.chunked = true
		r.HeaderMap.Set("Transfer-Encoding", "chunked")
	}

	// Status line
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.StatusCode, r.Status)
	_, _ = r.conn.Write([]byte(statusLine))

	// Headers
	_ = r.HeaderMap.Write(r.conn)

	// Header boundary CRLF
	_, _ = r.conn.Write([]byte("\r\n"))

	r.wroteHeader = true
}

func (r *Response) Write(data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, nil
	}

	if !r.wroteHeader {
		if r.Header().Get("Content-Type") == "" {
			r.Header().Set("Content-Type", http.DetectContentType(data))
		}
		r.WriteHeader(http.StatusOK)
	}

	// If using chunked encoding, wrap data in HTTP chunk format
	if r.chunked {
		chunkHeader := fmt.Sprintf("%x\r\n", len(data))
		if _, err := r.conn.Write([]byte(chunkHeader)); err != nil {
			return 0, err
		}
		if _, err := r.conn.Write(data); err != nil {
			return 0, err
		}
		if _, err := r.conn.Write([]byte("\r\n")); err != nil {
			return 0, err
		}
		return len(data), nil
	}

	// Standard write when Content-Length was pre-declared
	return r.conn.Write(data)
}

// Finish must be called after mux.ServeHTTP() returns to send the terminal chunk
func (r *Response) Finish() error {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	// Send final zero-length chunk to close the stream: "0\r\n\r\n"
	if r.chunked {
		_, err := r.conn.Write([]byte("0\r\n\r\n"))
		return err
	}
	return nil
}