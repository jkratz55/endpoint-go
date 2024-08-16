package endpoint

import (
	"time"
)

type TraceInfo struct {

	// The HTTP method of the request.
	Method string

	// The URL of the request.
	URL string

	// The HTTP status code of the response. This will be 0 if a response was not
	// received from the server.
	StatusCode int

	// The error that occurred while performing the request. This will be nil if
	// the Endpoint completed successfully.
	Error error

	// Duration it took to perform the DNS lookup.
	DNSLookup time.Duration

	// Duration it took to obtain a connection.
	ConnectionTime time.Duration

	// Duration the TLS handshake took.
	TLSHandshake time.Duration

	// Duration it took from the time the request was sent until the first byte
	// was received from the server.
	ServerTime time.Duration

	// Duration it took for the request end-to-end.
	TotalTime time.Duration

	// Indicates if the connection was reused
	ConnectionReused bool

	// Indicates if the connection was idle
	ConnectionIdle bool
}
