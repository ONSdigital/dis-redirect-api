package apierrors

import "errors"

// A list of error messages for Redirect API
var (
	ErrNegativeCount           = errors.New("the count must be a positive integer")
	ErrRedis                   = errors.New("redis returned an error")
	ErrInvalidCount            = errors.New("the count must be an integer giving the requested number of redirects")
	ErrInvalidOrNegativeCursor = errors.New("the redirects cursor was invalid. It must be a positive integer")
)
