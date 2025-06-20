package apierrors

import "errors"

// A list of error messages for Redirect API
var (
	ErrRedis = errors.New("redis returned an error")
)
