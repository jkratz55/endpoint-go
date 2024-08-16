package endpoint

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Middleware[T, R any] func(Endpoint[T, R]) Endpoint[T, R]

var (
	latency *prometheus.HistogramVec
	errs    *prometheus.CounterVec
)

// todo: lose http status codes :(
func Promtheus[T, R any](method, path string) Middleware[T, R] {
	return func(next Endpoint[T, R]) Endpoint[T, R] {
		return func(ctx context.Context, request T) (resp R, err error) {
			start := time.Now()
			defer func() {
				elapsed := time.Since(start)
				latency.WithLabelValues("POST", "/orders/v36").
					Observe(float64(elapsed.Milliseconds()))
				if err != nil {
					errs.WithLabelValues("POST", "/orders/v36").Inc()
				}
			}()

			return next(ctx, request)
		}
	}
}
