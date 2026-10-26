package main

import "github.com/firaslafi/http-from-scratch/internal/http1"


func main() {
	if err := http1.ListenAndServer(":8080", nil); err != nil {
		panic(err)
	}
}
