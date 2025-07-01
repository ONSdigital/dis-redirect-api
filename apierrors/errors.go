package apierrors

import "errors"

// A list of error messages for Redirect API
var (
	ErrRedis               = errors.New("redis returned an error")
	ErrInvalidNumRedirects = errors.New("the number of redirects to count was invalid")
	ErrInvalidCursor       = errors.New("the redirects cursor was invalid")
)
