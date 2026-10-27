package main

import (
	"fmt"
	"net/http"

	"github.com/firaslafi/http-from-scratch/internal/http1"
)


func main() {
	// using the default mux
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "response numro uno")
	})

	if err := http1.ListenAndServer(":8080", nil); err != nil {
		panic(err)
	}
}
