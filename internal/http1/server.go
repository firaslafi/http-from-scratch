package http1

import (
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
		log.Printf("connection accepted! %v", conn)
	}
}

func HandleConnection()  {
	
}
