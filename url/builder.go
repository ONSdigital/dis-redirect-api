package url

import (
	"fmt"
)

// Builder encapsulates the building of urls in a central place, with knowledge of the url structures and base host names.
type Builder struct {
	apiURL string
}

// NewBuilder returns a new instance of url.Builder
func NewBuilder(apiURL string) *Builder {
	return &Builder{
		apiURL: apiURL,
	}
}

// BuildRedirectSelfURL returns the self URL for a specific redirect id
func (builder *Builder) BuildRedirectSelfURL(redirectID string) string {
	return fmt.Sprintf("%s/v1/redirects/%s", builder.apiURL, redirectID)
}
