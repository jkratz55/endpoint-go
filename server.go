package endpoint

import (
	"context"
	"encoding/json"
	"net/http"
)

func Handler[T, R any](
	endpoint Endpoint[T, R],
	decoder DecodeRequestFunc[T],
	encoder EncodeResponseFunc[R]) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		req, err := decoder(ctx, r)
		if err != nil {
			// todo: handle error
		}

		response, err := endpoint(ctx, req)
		if err != nil {
			// todo: handle error
		}

		if err := encoder(ctx, w, response); err != nil {
			// todo: handle error
		}
	}

	return http.HandlerFunc(fn)
}

// ErrorHandler is a function that is invoked when an error occurs processing the
// client's request. The error handler is responsible for writing an appropriate
// response to the client.
type ErrorHandler func(ctx context.Context, w http.ResponseWriter, err error)

type Server[T, R any] struct {
	endpoint Endpoint[T, R]
	decoder  DecodeRequestFunc[T]
	encoder  EncodeResponseFunc[R]
}

func (s *Server[T, R]) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func EncodeJSONResponse[R any](ctx context.Context, w http.ResponseWriter, data R) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

func EncodeXMLResponse[R any](ctx context.Context, w http.ResponseWriter, data R) error {
	w.Header().Set("Content-Type", "application/xml")
	return json.NewEncoder(w).Encode(data)
}
