package endpoint

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrDecodeRequest  = errors.New("decode request")
	ErrEncodeRequest  = errors.New("encode request")
	ErrEncodeResponse = errors.New("encode response")
	ErrDecodeResponse = errors.New("decode response")
)

type ErrorResponse struct {
	Status    int    `json:"statusCode"`
	Message   string `json:"message"`
	Path      string `json:"path"`
	Details   any    `json:"details"`
	Timestamp int64  `json:"timestamp"`
}

type HttpError struct {
	Status int
	Header http.Header
	Body   []byte
}

func (he HttpError) Error() string {
	return fmt.Sprintf("http error: %d %s: %s", he.Status, http.StatusText(he.Status), he.Body)
}
