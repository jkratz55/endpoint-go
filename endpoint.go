package endpoint

import (
	"context"
	"net/http"
)

// Endpoint is a function type the represent a single endpoint in the server/client
// model. The endpoint is responsible for processing a request and returning a response.
type Endpoint[T, R any] func(ctx context.Context, request T) (R, error)

// EncodeRequestFunc is a function that is responsible for encoding the request
// data into the http.Request.
type EncodeRequestFunc[T any] func(ctx context.Context, r *http.Request, data T) error

// DecodeResponseFunc is a function that is responsible for decoding the response
// body into the response type.
type DecodeResponseFunc[R any] func(ctx context.Context, r *http.Response) (R, error)

// EncodeResponseFunc is a function that is responsible for encoding response data
// and sending it to the client.
type EncodeResponseFunc[R any] func(ctx context.Context, w http.ResponseWriter, data R) error

// DecodeRequestFunc is a function that is responsible for decoding a HTTP request
// into a type representing the request.
type DecodeRequestFunc[T any] func(ctx context.Context, r *http.Request) (T, error)

// CreateRequestFunc is a function that is responsible for creating a new
// http.Request.
type CreateRequestFunc[T any] func(ctx context.Context, request T) (*http.Request, error)
