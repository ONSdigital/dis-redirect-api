package models

// Redirect represents response body when retrieving a redirect
type Redirect struct {
	From  string        `json:"from,omitempty"`
	To    string        `json:"to,omitempty"`
	ID    string        `json:"id"`
	Links RedirectLinks `json:"links"`
}

// Redirects represents response body when retrieving a list of redirects
type Redirects struct {
	Count        int        `json:"count"`
	RedirectList []Redirect `json:"items"`
	Cursor       string     `json:"cursor"`
	NextCursor   string     `json:"next_cursor"`
	TotalCount   int        `json:"total_count"`
}

// RedirectLinks is a type that contains links relating to the individual redirect.
// Currently, it only contains one link, which is a link to itself.
type RedirectLinks struct {
	Self RedirectSelf `json:"self"`
}

// RedirectSelf represents a link to the individual redirect itself.
type RedirectSelf struct {
	Href string `json:"href"`
	ID   string `json:"id"`
}
