package endpoint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrorHandler is a function that is invoked when an error occurs processing the
// client's request. The error handler is responsible for writing an appropriate
// response to the client.
type ErrorHandler func(ctx context.Context, w http.ResponseWriter, err error)

type ServerHooks struct {

	// BeforeDecodeRequest is invoked before the request is decoded. It is not
	// safe to read the request body in this hook. If the request body is read
	// in this hook, it will need to be reset before it can be read again.
	BeforeDecodeRequest func(ctx context.Context, r *http.Request) context.Context

	// RequestDecoded is invoked after the request is decoded.
	RequestDecoded func(ctx context.Context, req interface{})

	// BeforeValidation is invoked before the request is validated.
	BeforeValidation func(ctx context.Context, req interface{})

	// RequestValidated is invoked after the request is validated.
	RequestValidated func(ctx context.Context, ok bool, violations []ValidationViolation)

	// AfterEndpoint is invoked after the endpoint is invoked.
	AfterEndpoint func(ctx context.Context, w http.ResponseWriter)

	// Invoked after the ServeHTTP method has completed execution.
	Finalizer func(ctx context.Context, statusCode int, r *http.Request)
}

type Server[T, R any] struct {
	endpoint     Endpoint[T, R]
	decoder      DecodeRequestFunc[T]
	encoder      EncodeResponseFunc[R]
	validator    Validator[T]
	errorHandler ErrorHandler
	hooks        ServerHooks
}

func NewServer[T, R any](
	e Endpoint[T, R],
	d DecodeRequestFunc[T],
	enc EncodeResponseFunc[R]) *Server[T, R] {

	s := &Server[T, R]{
		endpoint:     e,
		decoder:      d,
		encoder:      enc,
		errorHandler: defaultServerErrorHandler(),
	}

	// todo: allow opts/hooks

	return s
}

func (s *Server[T, R]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if s.hooks.Finalizer != nil {
		rw := &responseWriter{
			w:       w,
			code:    200,
			written: 0,
		}
		defer func() {
			s.hooks.Finalizer(ctx, rw.code, r)
		}()
		w = rw
	}

	if s.hooks.BeforeDecodeRequest != nil {
		ctx = s.hooks.BeforeDecodeRequest(ctx, r)
	}

	req, err := s.decoder(ctx, r)
	if err != nil {
		decodeErr := fmt.Errorf("%w: %w", ErrDecodeRequest, err)
		s.errorHandler(ctx, w, decodeErr)
		return
	}

	if s.hooks.RequestDecoded != nil {
		s.hooks.RequestDecoded(ctx, req)
	}

	// Validator hook is only called if one is set. If the validator returns false
	// indicating the request isn't valid we'll send a 400 Bad Request response to
	// the client and never bother invoking the endpoint.
	if s.validator != nil {

		if s.hooks.BeforeValidation != nil {
			s.hooks.BeforeValidation(ctx, req)
		}

		ok, violations := s.validator(req)

		if s.hooks.RequestValidated != nil {
			s.hooks.RequestValidated(ctx, ok, violations)
		}

		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Status:    http.StatusBadRequest,
				Message:   "Bad Request",
				Path:      r.URL.Path,
				Details:   violations,
				Timestamp: time.Now().Unix(),
			})
			return
		}
	}

	response, err := s.endpoint(ctx, req)
	if err != nil {
		s.errorHandler(ctx, w, err)
		return
	}

	if s.hooks.AfterEndpoint != nil {
		s.hooks.AfterEndpoint(ctx, w)
	}

	if err := s.encoder(ctx, w, response); err != nil {
		s.errorHandler(ctx, w, fmt.Errorf("%w: %w", ErrEncodeResponse, err))
		return
	}
}

func defaultServerErrorHandler() ErrorHandler {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		switch {
		case errors.Is(err, ErrDecodeRequest):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Status:    http.StatusBadRequest,
				Message:   "Bad Request",
				Details:   "Server was unable to parse or unmarshall the request",
				Timestamp: time.Now().Unix(),
			})
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Status:    http.StatusInternalServerError,
				Message:   "Internal Server Error",
				Details:   err.Error(),
				Timestamp: time.Now().Unix(),
			})
		}
	}
}

func EncodeJSONResponse(ctx context.Context, w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(data)
}
