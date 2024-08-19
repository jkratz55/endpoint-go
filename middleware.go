package endpoint

// Middleware is a function that accepts an Endpoint and returns an Endpoint.
// Middleware wraps an endpoint allowing for additional processing before and/or
// after the endpoint is invoked.
type Middleware[T, R any] func(Endpoint[T, R]) Endpoint[T, R]
