package sdk

import (
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-net/v3/request"
)

const (
	// List of available headers
	Authorization string = request.AuthHeaderKey
)

// Options is a struct containing for customised options for the API client
type Options struct {
	Headers http.Header
	Query   url.Values
}

func setHeaders(req *http.Request, headers http.Header) {
	for name, values := range headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
}
