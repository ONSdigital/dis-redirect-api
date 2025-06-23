package models

// Redirect represents response body when retrieving a redirect
type Redirect struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}
