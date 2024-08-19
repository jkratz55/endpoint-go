package endpoint

import (
	"net/http"
)

type responseWriter struct {
	w       http.ResponseWriter
	code    int
	written int64
}

func (r responseWriter) Header() http.Header {
	return r.w.Header()
}

func (r responseWriter) Write(bytes []byte) (int, error) {
	n, err := r.w.Write(bytes)
	r.written += int64(n)
	return n, err
}

func (r responseWriter) WriteHeader(statusCode int) {
	r.code = statusCode
	r.w.WriteHeader(statusCode)
}
