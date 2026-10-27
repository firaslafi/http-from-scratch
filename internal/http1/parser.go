package http1

import (
	"bufio"
	"net/http"
)

func ParseRequest(reader *bufio.Reader) (*http.Request, error) {
	panic("parse req is not implemented")
}