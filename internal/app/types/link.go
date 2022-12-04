package types

// Link represents a pair of short and original urls for GetUserStorage handler.
type Link struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// BatchLink represents a link for batch requests.
type BatchLink struct {
	CorrelationID string
	ShortURL      string
	OriginalURL   string
}

// OriginalLink represents an original link and current state.
type OriginalLink struct {
	OriginalURL string
	Deleted     bool
}

// BatchLinks represents a slice of links for batch requests.
type BatchLinks []BatchLink
