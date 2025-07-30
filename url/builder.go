package url

import (
	"fmt"
	"net/url"
)

// Builder encapsulates the building of urls in a central place, with knowledge of the url structures and base host names.
type Builder struct {
	redirectAPIURL *url.URL
}

// NewBuilder returns a new instance of url.Builder
func NewBuilder(redirectAPIURL *url.URL) *Builder {
	return &Builder{
		redirectAPIURL: redirectAPIURL,
	}
}

func (builder *Builder) GetRedirectAPIURL() *url.URL {
	return builder.redirectAPIURL
}

// BuildRedirectSelfURL returns the self URL for a specific redirect id
func (builder *Builder) BuildRedirectSelfURL(redirectID string) string {
	return fmt.Sprintf("%s%s", builder.redirectAPIURL.String(), redirectID)
}
