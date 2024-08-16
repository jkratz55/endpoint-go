package endpoint

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"
)

// HttpClient is an interface that defines a type capable of making HTTP requests.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client ...
type Client[T, R any] struct {
	client           HttpClient
	reqBuilder       CreateRequestFunc[T]
	decoder          DecodeResponseFunc[R]
	requestPrepared  []RequestPrepared
	responseReceived []ResponseReceived
	errFunc          OnError
	traceEnabled     bool
}

func NewClient[T, R any](
	method string,
	uri *url.URL,
	encoder EncodeRequestFunc[T],
	decoder DecodeResponseFunc[R],
	opts ...ClientOptions[T, R]) *Client[T, R] {

	client := &Client[T, R]{
		client:           http.DefaultClient,
		reqBuilder:       makeCreateRequest(method, uri, encoder),
		decoder:          decoder,
		requestPrepared:  make([]RequestPrepared, 0),
		responseReceived: make([]ResponseReceived, 0),
		errFunc:          func(ctx context.Context, err error) {},
		traceEnabled:     false,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client[T, R]) Endpoint() Endpoint[T, R] {
	return func(ctx context.Context, request T) (R, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var (
			zero R

			// Timestamps for httptrace
			getConnStart time.Time
			dnsStart     time.Time
			connectStart time.Time
			tlsStart     time.Time
			reqSent      time.Time
			start        time.Time

			// Durations for httptrace
			getConnDur time.Duration
			connDur    time.Duration
			dnsDur     time.Duration
			tlsDur     time.Duration
			serverDur  time.Duration
		)
		start = time.Now()

		req, err := c.reqBuilder(ctx, request)
		if err != nil {
			return zero, err
		}

		if c.traceEnabled {
			clientTrace := &httptrace.ClientTrace{
				GetConn: func(_ string) {
					getConnStart = time.Now()
				},
				GotConn: func(info httptrace.GotConnInfo) {
					getConnDur = time.Since(getConnStart)
				},
				GotFirstResponseByte: func() {
					serverDur = time.Since(reqSent)
				},
				DNSStart: func(info httptrace.DNSStartInfo) {
					dnsStart = time.Now()
				},
				DNSDone: func(info httptrace.DNSDoneInfo) {
					dnsDur = time.Since(dnsStart)
				},
				ConnectStart: func(network, addr string) {
					connectStart = time.Now()
				},
				ConnectDone: func(network, addr string, err error) {
					connDur = time.Since(connectStart)
				},
				TLSHandshakeStart: func() {
					tlsStart = time.Now()
				},
				TLSHandshakeDone: func(state tls.ConnectionState, err error) {
					tlsDur = time.Since(tlsStart)
				},
				WroteRequest: func(info httptrace.WroteRequestInfo) {
					reqSent = time.Now()
				},
			}
			ctx = httptrace.WithClientTrace(ctx, clientTrace)
		}

		for _, fn := range c.requestPrepared {
			ctx = fn(ctx, req)
		}

		resp, err := c.client.Do(req.WithContext(ctx))
		if err != nil {
			return zero, err
		}

		for _, fn := range c.responseReceived {
			ctx = fn(ctx, resp)
		}

		response, err := c.decoder(ctx, resp)
		if err != nil {
			return zero, err
		}

		totalDur := time.Since(start)

		if c.traceEnabled {
			code := 0
			if resp != nil {
				code = resp.StatusCode
			}
			info := TraceInfo{
				Method:           req.Method,
				URL:              req.URL.String(),
				StatusCode:       code,
				Error:            err,
				DNSLookup:        dnsDur,
				ConnectionTime:   connDur,
				TLSHandshake:     tlsDur,
				ServerTime:       serverDur,
				TotalTime:        totalDur,
				ConnectionReused: false,
				ConnectionIdle:   false,
			}
			fmt.Println(info)
		}

		return response, nil
	}
}

func EncodeJSONRequest[T any](_ context.Context, r *http.Request, data T) error {
	r.Header.Set("Content-Type", "application/json")
	var buf bytes.Buffer
	r.Body = io.NopCloser(&buf)
	return json.NewEncoder(&buf).Encode(data)
}

func EncodeXMLRequest[T any](_ context.Context, r *http.Request, data T) error {
	r.Header.Set("Content-Type", "application/xml")
	var buf bytes.Buffer
	r.Body = io.NopCloser(&buf)
	return xml.NewEncoder(&buf).Encode(data)
}

func NopRequestEncoder[T any](_ context.Context, _ *http.Request, _ T) error {
	return nil
}

func makeCreateRequest[T any](method string, uri *url.URL, encoder EncodeRequestFunc[T]) CreateRequestFunc[T] {
	return func(ctx context.Context, request T) (*http.Request, error) {
		req, err := http.NewRequest(method, uri.String(), nil)
		if err != nil {
			return nil, err
		}

		if err := encoder(ctx, req, request); err != nil {
			return nil, err
		}

		return req, nil
	}
}
