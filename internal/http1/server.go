package http1

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
)

// listen in TCP net addr and calls Serve() with the handler on conns
// accepted conns configured to enable TCP keep-alive
func ListenAndServer(addr string, handler http.Handler) error {
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	defer listener.Close()

	// if hanlder is nil, then use defaultmux
	if handler == nil {
		handler = http.DefaultServeMux
	}
	log.Printf("started listening on %s", listener.Addr().String())
	for {
		conn, err := listener.Accept()

		if err != nil {
			return fmt.Errorf("accept: %w", err)
		}
		// log.Printf("connection accepted! %v", conn)
		go HandleConnection(conn, handler)
	}
}

// func HandleConnection(conn net.Conn, mux http.Handler) error 
func HandleConnection(conn net.Conn, mux http.Handler) error {
	defer conn.Close()



	// read data from connection and parse it (sepreate func)
	buff := bufio.NewReader(conn)

	// change this to my own implementation
	request, err := http.ReadRequest(buff)
	
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}

	// fmt.Printf("request:\n%v+", request)

	response := NewResponse(conn)
	response.Request = request

	response.Header().Add("testkey", "testvalue")
	fmt.Printf("\nresponse header:\n%v", response.Header())

	mux.ServeHTTP(response, request)

	
	// if err != nil {
	// 	return fmt.Errorf("writing response: %w", err)
	// }

	if !response.wroteHeader {
		response.WriteHeader(http.StatusOK)
	}

	return nil

}
