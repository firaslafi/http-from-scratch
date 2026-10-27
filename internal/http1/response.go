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

func (r *Response) Write(data []byte) (n int, err error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.conn.Write(data)
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
	// Write HTTP status line
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.StatusCode, r.Status)
	_, _ = r.conn.Write([]byte(statusLine))

	// Write all headers
	_ = r.HeaderMap.Write(r.conn)

	// Write CRLF to mark the boundary between headers and body
	_, _ = r.conn.Write([]byte("\r\n"))

	r.wroteHeader = true
}
