package http1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func ListenAndServe(addr string, handler http.Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	if handler == nil {
		handler = http.DefaultServeMux
	}

	log.Printf("Custom HTTP server running on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go func(c net.Conn) {
			if err := HandleConnection(c, handler); err != nil {
				log.Printf("[%s] connection error: %v", c.RemoteAddr(), err)
			}
		}(conn)
	}
}

func HandleConnection(conn net.Conn, mux http.Handler) error {
	defer conn.Close()

	bufr := bufio.NewReader(conn)

	for {
		// Parse incoming request (blocks indefinitely until bytes arrive or client closes)
		req, err := ReadRequest(bufr)
		if err != nil {
			// Clean disconnect by client is normal termination, not an error
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read request: %w", err)
		}
		// Reject unsupported HTTP versions
		if req.ProtoMajor != 1 {
			res := NewResponse(conn)
			res.Header().Set("Connection", "close")
			res.WriteHeader(http.StatusHTTPVersionNotSupported)
			_ = res.Finish()
			return nil
		}

		// Negotiate Keep-Alive state
		keepAlive := shouldKeepAlive(req)

		res := NewResponse(conn)
		if !keepAlive {
			res.Header().Set("Connection", "close")
		} else if isHTTP10(req) {
			res.Header().Set("Connection", "keep-alive")
		}

		// Dispatch to Application Handler
		mux.ServeHTTP(res, req)

		// Complete response cycle (flush headers/body/terminal chunk)
		if err := res.Finish(); err != nil {
			return fmt.Errorf("finish response: %w", err)
		}

		// Drain any remaining unread body to align bufr with the next request
		if req.Body != nil {
			_, _ = io.Copy(io.Discard, req.Body)
			_ = req.Body.Close()
		}

		// Terminate connection if closing was requested
		if !keepAlive || strings.EqualFold(res.Header().Get("Connection"), "close") {
			return nil
		}
	}
}

func shouldKeepAlive(req *http.Request) bool {
	connVal := strings.ToLower(strings.TrimSpace(req.Header.Get("Connection")))

	// HTTP/1.0 closes by default unless "keep-alive" is explicitly requested
	if isHTTP10(req) {
		return strings.Contains(connVal, "keep-alive")
	}

	// HTTP/1.1+ defaults to keep-alive unless "close" is specified
	return !strings.Contains(connVal, "close")
}

func isHTTP10(req *http.Request) bool {
	return req.Proto == "HTTP/1.0" || (req.ProtoMajor == 1 && req.ProtoMinor == 0)
}