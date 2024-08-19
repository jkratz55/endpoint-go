package endpoint

type ClientOption[T, R any] func(*Client[T, R])

func WithClient[T, R any](client HttpClient) ClientOption[T, R] {
	return func(c *Client[T, R]) {
		c.client = client
	}
}

func WithRequestBuilder[T, R any](fn CreateRequestFunc[T]) ClientOption[T, R] {
	return func(c *Client[T, R]) {
		c.reqBuilder = fn
	}
}

func WithClientHooks[T, R any](hooks ClientHooks) ClientOption[T, R] {
	return func(c *Client[T, R]) {
		c.hooks = hooks
	}
}

type ServerOption[T, R any] func(*Server[T, R])

func WithServerHooks[T, R any](hooks ServerHooks) ServerOption[T, R] {
	return func(s *Server[T, R]) {
		s.hooks = hooks
	}
}

func WithServerValidator[T any](fn Validator[T]) ServerOption[T, T] {
	return func(s *Server[T, T]) {
		s.validator = fn
	}
}

func WithServerErrorHandler[T, R any](fn ErrorHandler) ServerOption[T, R] {
	return func(s *Server[T, R]) {
		s.errorHandler = fn
	}
}
