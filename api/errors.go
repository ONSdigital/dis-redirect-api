package api

import "errors"

// A list of error messages for Redirect API
var (
	ErrNegativeCount           = errors.New("the count must be a positive integer")
	ErrInternal                = errors.New("internal error")
	ErrInvalidCount            = errors.New("the count must be an integer giving the requested number of redirects")
	ErrInvalidOrNegativeCursor = errors.New("the redirects cursor was invalid. It must be a positive integer")
	ErrInvalidBase64Id         = errors.New("the base64 id provided is invalid")
	ErrNotFound                = errors.New("not found")
	ErrInvalidRequestBody      = errors.New("the request body provided is invalid")
	ErrIdFromMismatch          = errors.New("the 'from' field does not match the base64 id")
	ErrFromToNotRelative       = errors.New("'from' and 'to' must be relative paths starting with '/'")
	ErrCircularPaths           = errors.New("'from' and 'to' cannot be the same")
)
