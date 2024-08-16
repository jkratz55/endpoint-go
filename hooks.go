package endpoint

import (
	"context"
	"net/http"
	"time"
)

// RequestPrepared is a function that is invoked after the request is prepared
// but before it is sent to the server. The request can safely be modified in
// this hook.
type RequestPrepared func(context.Context, *http.Request) context.Context

// todo: Maybe not possible?
// RequestSent is a function that is invoked after the request is sent to the
// server. At this point the request should not be modified as it has already
// been sent to the server.
// type RequestSent func(context.Context, *http.Request) context.Context

// ResponseReceived is a function that is invoked after the response is received
// from the server but before the DecodeResponseFunc is invoked. ResponseReceived
// will not be invoked if the server never responds. Common reasons for this include
// network errors, timeouts, etc.
type ResponseReceived func(context.Context, *http.Response) context.Context

// OnError is a function that is invoked when an error occurs and a response
// is not received from the server.
type OnError func(context.Context, error)

// OnComplete is a function hook that is invoked after the Endpoint has completed
// execution.
//
// If a response was not received from the server, the error will be non-nil
// and the statusCode will be 0.
type OnComplete func(ctx context.Context, statusCode int, dur time.Duration, err error)
