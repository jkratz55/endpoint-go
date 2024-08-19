package endpoint

// Validator is a function hook that is invoked after the request is decoded but
// before the endpoint is invoked. The validator is responsible for validating the
// request to ensure it is valid before the endpoint is invoked.
type Validator[T any] func(T any) (bool, []ValidationViolation)

// ValidationViolation represents a validation error that occurred while validating
// a client request.
type ValidationViolation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
