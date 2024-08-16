package endpoint

import (
	"context"
	"net/http"
)

type Endpoint[T, R any] func(ctx context.Context, request T) (R, error)

type EncodeRequestFunc[T any] func(ctx context.Context, r *http.Request, data T) error

type DecodeResponseFunc[R any] func(ctx context.Context, r *http.Response) (R, error)

type EncodeResponseFunc[R any] func(ctx context.Context, w http.ResponseWriter, data R) error

type DecodeRequestFunc[T any] func(ctx context.Context, r *http.Request) (T, error)

type CreateRequestFunc[T any] func(ctx context.Context, request T) (*http.Request, error)

func Nop(_ context.Context, _ interface{}) (interface{}, error) {
	return nil, nil
}
