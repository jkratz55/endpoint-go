package endpoint

type ClientOptions[T, R any] func(*Client[T, R])

func WithClient[T, R any](client HttpClient) ClientOptions[T, R] {
	return func(c *Client[T, R]) {
		c.client = client
	}
}

func OnRequestPrepared[T, R any](fn RequestPrepared) ClientOptions[T, R] {
	return func(c *Client[T, R]) {
		c.requestPrepared = append(c.requestPrepared, fn)
	}
}

// func OnRequestSent[T, R any](fn ) ClientOptions[T, R] {
// 	return func(c *Client[T, R]) {
// 		c.requestSent = append(c.requestSent, fn)
// 	}
// }

func OnResponseReceived[T, R any](fn ResponseReceived) ClientOptions[T, R] {
	return func(c *Client[T, R]) {
		c.responseReceived = append(c.responseReceived, fn)
	}
}

func OnErrorHandler[T, R any](fn OnError) ClientOptions[T, R] {
	return func(c *Client[T, R]) {
		c.errFunc = fn
	}
}

func WithRequestBuilder[T, R any](fn CreateRequestFunc[T]) ClientOptions[T, R] {
	return func(c *Client[T, R]) {
		c.reqBuilder = fn
	}
}
