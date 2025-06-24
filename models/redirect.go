package models

// Redirect represents response body when retrieving a redirect
type Redirect struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}
