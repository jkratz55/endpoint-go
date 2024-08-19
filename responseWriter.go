package endpoint

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

type responseWriter struct {
	w       http.ResponseWriter
	code    int
	written int64
}

func (r *responseWriter) Header() http.Header {
	return r.w.Header()
}

func (r *responseWriter) Write(bytes []byte) (int, error) {
	n, err := r.w.Write(bytes)
	r.written += int64(n)
	return n, err
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.code = statusCode
	r.w.WriteHeader(statusCode)
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.w.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("hijacking not supported: wrapped ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

func (r *responseWriter) Flush() {
	if flusher, ok := r.w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := r.w.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}
