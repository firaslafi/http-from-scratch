package main

import (
	"fmt"
	"net/http"

	"github.com/firaslafi/http-from-scratch/internal/http1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const addr = ":8080"

func main() {
	mux := http1.New()
	mux.HandleFunc(http.MethodGet, "/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	})

	mux.HandleFunc(http.MethodGet, "/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello from custom HTTP/1.1 server!")
	})

	mux.HandleFunc(http.MethodGet, "/metrics", promhttp.Handler().ServeHTTP)

	if err := http1.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}
}
