package endpoint

type ClientOptions[T, R any] func(*Client[T, R])

func WithClient[T, R any](client HttpClient) ClientOptions[T, R] {
	return func(c *Client[T, R]) {
		c.client = client
	}
}

func WithRequestBuilder[T, R any](fn CreateRequestFunc[T]) ClientOptions[T, R] {
	return func(c *Client[T, R]) {
		c.reqBuilder = fn
	}
}

func WithClientHooks[T, R any](hooks ClientHooks) ClientOptions[T, R] {
	return func(c *Client[T, R]) {
		c.hooks = hooks
	}
}
