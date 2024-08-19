package endpoint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HttpClient is an interface that defines a type capable of making HTTP requests.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientHooks is a type for defining hooks that are invoked during the lifecycle
// of an Endpoint.
type ClientHooks struct {

	// BeforePrepareRequest is invoked before the request is prepared/initialized.
	BeforePrepareRequest func(ctx context.Context)

	// RequestPrepared is invoked after the request is prepared/initialized, but
	// before it is sent to the server. The request can safely be modified.
	RequestPrepared func(context.Context, *http.Request) context.Context

	// BeforeSendRequest is invoked before the request is sent to the server.
	BeforeSendRequest func(ctx context.Context)

	// ResponseReceived is invoked after the response is received from the server
	// but before the DecodeResponseFunc is invoked. ResponseReceived will not be
	// invoked if the server never responds. Common reasons for this include network
	// errors, timeouts, etc.
	ResponseReceived func(context.Context, *http.Response) context.Context

	// AfterDecodeResponse is invoked after the response is decoded.
	ResponseDecoded func(ctx context.Context)

	// Finalizer is invoked after the Endpoint has completed execution.
	//
	// If a response was not received from the server, the statusCode will be 0.
	Finalizer func(ctx context.Context, statusCode int, err error)

	// OnError is invoked when an error is returned by the implementation of the
	// HttpClient interface.
	OnError func(context.Context, error)
}

// Client is a type for building an Endpoint invoke a remote service over HTTP.
type Client[T, R any] struct {
	client     HttpClient
	reqBuilder CreateRequestFunc[T]
	decoder    DecodeResponseFunc[R]
	hooks      ClientHooks
}

// NewClient initializes a new Client which acts as a builder for an Endpoint.
func NewClient[T, R any](
	method string,
	uri string,
	encoder EncodeRequestFunc[T],
	decoder DecodeResponseFunc[R],
	opts ...ClientOptions[T, R]) *Client[T, R] {

	client := &Client[T, R]{
		client:     http.DefaultClient,
		reqBuilder: makeCreateRequest(method, uri, encoder),
		decoder:    decoder,
		hooks:      ClientHooks{},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// NewCustomRequestClient initializes a new Client with a custom
// CreateRequestFunc. This gives the caller more control over how the request
// are built.
func NewCustomRequestClient[T, R any](
	reqBuilder CreateRequestFunc[T],
	decoder DecodeResponseFunc[R],
	opts ...ClientOptions[T, R]) *Client[T, R] {

	client := &Client[T, R]{
		client:     http.DefaultClient,
		reqBuilder: reqBuilder,
		decoder:    decoder,
		hooks:      ClientHooks{},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Endpoint returns an Endpoint that can be used to invoke the remote service.
func (c *Client[T, R]) Endpoint() Endpoint[T, R] {
	return func(ctx context.Context, request T) (R, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var (
			zero       R
			statusCode int
			err        error
		)

		if c.hooks.Finalizer != nil {
			defer func() {
				c.hooks.Finalizer(ctx, statusCode, err)
			}()
		}

		if c.hooks.BeforePrepareRequest != nil {
			c.hooks.BeforePrepareRequest(ctx)
		}

		req, err := c.reqBuilder(ctx, request)
		if err != nil {
			return zero, err
		}

		if c.hooks.RequestPrepared != nil {
			ctx = c.hooks.RequestPrepared(ctx, req)
		}

		if c.hooks.BeforeSendRequest != nil {
			c.hooks.BeforeSendRequest(ctx)
		}

		resp, err := c.client.Do(req.WithContext(ctx))
		if err != nil {
			return zero, err
		}
		statusCode = resp.StatusCode

		if c.hooks.ResponseReceived != nil {
			ctx = c.hooks.ResponseReceived(ctx, resp)
		}

		response, err := c.decoder(ctx, resp)
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrDecodeResponse, err)
			if c.hooks.OnError != nil {
				c.hooks.OnError(ctx, err)
			}
			return zero, err
		}

		if c.hooks.ResponseDecoded != nil {
			c.hooks.ResponseDecoded(ctx)
		}

		return response, nil
	}
}

// EncodeJSONRequest is a EncodeRequestFunc that encodes the request as JSON.
func EncodeJSONRequest(_ context.Context, r *http.Request, data interface{}) error {
	r.Header.Set("Content-Type", "application/json")
	var buf bytes.Buffer
	r.Body = io.NopCloser(&buf)
	return json.NewEncoder(&buf).Encode(data)
}

// NopRequestEncoder is a EncodeRequestFunc that does nothing and always returns
// nil.
func NopRequestEncoder(_ context.Context, _ *http.Request, _ interface{}) error {
	return nil
}

// DecodeJSONResponse is a DecodeResponseFunc that decodes the response as JSON
// into type R if the status code from the server is considered successful (2XX).
// If the status code is not 2XX an HttpError is returned.
func DecodeJSONResponse[R any](_ context.Context, resp *http.Response) (R, error) {
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		var data R
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return data, err
		}
		return data, nil
	}

	body, _ := io.ReadAll(resp.Body)
	return *new(R), HttpError{
		Status: resp.StatusCode,
		Header: resp.Header,
		Body:   body,
	}
}

func makeCreateRequest[T any](method string, uri string, encoder EncodeRequestFunc[T]) CreateRequestFunc[T] {
	return func(ctx context.Context, request T) (*http.Request, error) {
		req, err := http.NewRequest(method, uri, nil)
		if err != nil {
			return nil, err
		}

		if err := encoder(ctx, req, request); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrEncodeRequest, err)
		}

		return req, nil
	}
}
