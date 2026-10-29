package main

import (
	"fmt"
	"net/http"

	"github.com/firaslafi/http-from-scratch/internal/http1"
)

const addr = ":8080"

func main() {
	mux := http1.New()
	mux.HandleFunc(http.MethodGet, "/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	})

	if err := http1.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}
}
